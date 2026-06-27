# ── Default VPC + Subnets ─────────────────────────────────────
data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

# ── Launch Template for MD Workers ───────────────────────────
resource "aws_launch_template" "worker_md" {
  name_prefix   = "${var.project_name}-worker-md-"
  image_id      = "ami-0f58b397bc5c1f2e8" # Ubuntu 22.04 LTS ap-south-1
  instance_type = "t3.micro"              # free tier in ap-south-1; use c5.2xlarge in prod

  iam_instance_profile {
    name = aws_iam_instance_profile.ec2_profile.name
  }

  vpc_security_group_ids = [aws_security_group.app.id]

  # SSM fetch + build-from-source bootstrap
  # Replace the git clone + go build block with an ECR pull once images are pushed:
  #   aws ecr get-login-password | docker login --username AWS --password-stdin $ECR
  #   docker pull $ECR/docplatform/worker-md:latest
  user_data = base64encode(<<-EOF
    #!/bin/bash
    set -e
    exec > /var/log/worker-init.log 2>&1

    # Swap space — required for Go builds on t2.micro (1 GB RAM)
    fallocate -l 2G /swapfile
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile

    # System packages
    apt-get update -y
    apt-get install -y awscli git wget \
        wkhtmltopdf libxrender1 libxext6 libfontconfig1 \
        ca-certificates

    # Install Go 1.25
    wget -q https://go.dev/dl/go1.25.3.linux-amd64.tar.gz -O /tmp/go.tar.gz
    tar -C /usr/local -xf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    export PATH=$PATH:/usr/local/go/bin

    # Resolve current region from instance metadata
    REGION=$(curl -s http://169.254.169.254/latest/meta-data/placement/region)

    # Fetch configuration from SSM Parameter Store
    fetch() {
      aws ssm get-parameter --name "$1" \
          --with-decryption \
          --query 'Parameter.Value' \
          --output text \
          --region "$REGION"
    }

    S3_BUCKET=$(fetch /docplatform/s3-bucket)
    DYNAMODB_TABLE=$(fetch /docplatform/dynamodb-table)
    SQS_QUEUE_URL=$(fetch /docplatform/sqs-queue-url)
    THUMB_SQS_QUEUE_URL=$(fetch /docplatform/thumb-sqs-queue-url)

    # Write env file (never logs secrets to stdout)
    cat > /opt/worker-md.env <<ENVEOF
    AWS_REGION=$REGION
    S3_BUCKET=$S3_BUCKET
    DYNAMODB_TABLE=$DYNAMODB_TABLE
    SQS_QUEUE_URL=$SQS_QUEUE_URL
    THUMB_SQS_QUEUE_URL=$THUMB_SQS_QUEUE_URL
    ENVEOF
    chmod 600 /opt/worker-md.env

    # Clone full repo so go.work is present at the root
    git clone https://github.com/Vedanttamboli17/mdtopdf-devops /opt/docplatform

    # Build — Go walks up the directory tree and finds go.work at /opt/docplatform/
    cd /opt/docplatform/worker-md
    /usr/local/go/bin/go build -ldflags="-w -s" -o /usr/local/bin/worker-md .

    # Systemd service
    cat > /etc/systemd/system/worker-md.service <<SVCEOF
    [Unit]
    Description=DocPlatform MD Worker
    After=network.target
    StartLimitIntervalSec=0

    [Service]
    EnvironmentFile=/opt/worker-md.env
    ExecStart=/usr/local/bin/worker-md
    Restart=always
    RestartSec=10
    StandardOutput=journal
    StandardError=journal

    [Install]
    WantedBy=multi-user.target
    SVCEOF

    systemctl daemon-reload
    systemctl enable worker-md
    systemctl start worker-md
  EOF
  )

  tag_specifications {
    resource_type = "instance"
    tags = {
      Name    = "${var.project_name}-worker-md"
      Project = var.project_name
      Role    = "worker-md"
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}

# ── Auto Scaling Group ─────────────────────────────────────────
resource "aws_autoscaling_group" "worker_md" {
  name                = "${var.project_name}-worker-md-asg"
  min_size            = 1
  max_size            = 5
  desired_capacity    = 1
  vpc_zone_identifier = data.aws_subnets.default.ids

  launch_template {
    id      = aws_launch_template.worker_md.id
    version = "$Latest"
  }

  # Wait for instances to pass health check before marking healthy
  health_check_type         = "EC2"
  health_check_grace_period = 300

  tag {
    key                 = "Name"
    value               = "${var.project_name}-worker-md"
    propagate_at_launch = true
  }

  tag {
    key                 = "Project"
    value               = var.project_name
    propagate_at_launch = true
  }
}

# ── Scaling Policies ──────────────────────────────────────────
resource "aws_autoscaling_policy" "worker_md_scale_out" {
  name                   = "${var.project_name}-worker-md-scale-out"
  autoscaling_group_name = aws_autoscaling_group.worker_md.name
  adjustment_type        = "ChangeInCapacity"
  scaling_adjustment     = 1
  cooldown               = 120 # 2 min: allow new instance to start before another triggers
}

resource "aws_autoscaling_policy" "worker_md_scale_in" {
  name                   = "${var.project_name}-worker-md-scale-in"
  autoscaling_group_name = aws_autoscaling_group.worker_md.name
  adjustment_type        = "ChangeInCapacity"
  scaling_adjustment     = -1
  cooldown               = 300 # 5 min: match the scale-in alarm period
}

# ── CloudWatch Alarms ─────────────────────────────────────────
resource "aws_cloudwatch_metric_alarm" "md_queue_high" {
  alarm_name          = "${var.project_name}-md-queue-high"
  alarm_description   = "Scale out MD workers when queue depth exceeds 10"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 60
  statistic           = "Average"
  threshold           = 10
  treat_missing_data  = "notBreaching"

  dimensions = {
    QueueName = aws_sqs_queue.main.name
  }

  alarm_actions = [aws_autoscaling_policy.worker_md_scale_out.arn]

  tags = { Project = var.project_name }
}

resource "aws_cloudwatch_metric_alarm" "md_queue_low" {
  alarm_name          = "${var.project_name}-md-queue-low"
  alarm_description   = "Scale in MD workers when queue depth stays below 2 for 5 min"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 300 # 5-minute window — don't scale in too aggressively
  statistic           = "Average"
  threshold           = 2
  treat_missing_data  = "notBreaching"

  dimensions = {
    QueueName = aws_sqs_queue.main.name
  }

  alarm_actions = [aws_autoscaling_policy.worker_md_scale_in.arn]

  tags = { Project = var.project_name }
}
