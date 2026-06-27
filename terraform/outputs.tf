output "ec2_public_ip" {
  description = "SSH and browser address of your server"
  value       = aws_instance.app.public_ip
}

output "s3_bucket_name" {
  value = aws_s3_bucket.main.bucket
}

output "dynamodb_table_name" {
  value = aws_dynamodb_table.jobs.name
}

output "sqs_queue_url" {
  description = "Paste this in .env file"
  value       = aws_sqs_queue.main.url
}

output "docx_sqs_queue_url" {
  value = aws_sqs_queue.docx_queue.url
}

output "thumb_sqs_queue_url" {
  value = aws_sqs_queue.thumb_queue.url
}

output "notification_sqs_queue_url" {
  value = aws_sqs_queue.notification_queue.url
}

output "tenants_dynamodb_table_name" {
  value = aws_dynamodb_table.tenants.name
}

output "worker_md_asg_name" {
  description = "Name of the MD worker Auto Scaling Group"
  value       = aws_autoscaling_group.worker_md.name
}

output "ssm_parameter_paths" {
  description = "SSM parameter paths for worker bootstrap — fetch with: aws ssm get-parameter --name <path> --with-decryption"
  value = {
    aws_region                 = aws_ssm_parameter.aws_region.name
    s3_bucket                  = aws_ssm_parameter.s3_bucket.name
    dynamodb_table             = aws_ssm_parameter.dynamodb_table.name
    sqs_queue_url              = aws_ssm_parameter.sqs_queue_url.name
    docx_sqs_queue_url         = aws_ssm_parameter.docx_sqs_queue_url.name
    thumb_sqs_queue_url        = aws_ssm_parameter.thumb_sqs_queue_url.name
    notification_sqs_queue_url = aws_ssm_parameter.notification_sqs_queue_url.name
  }
}