# Built from the REPO ROOT so the workspace-shared module is in context:
#   docker build -t docplatform-worker-md:latest -f worker-md/Dockerfile .
#
# ── Stage 1: Build ─────────────────────────────────────────────
FROM golang:1.25.3 AS builder

ENV GOWORK=off
WORKDIR /src

COPY shared/go.mod shared/go.sum ./shared/
COPY worker-md/go.mod worker-md/go.sum ./worker-md/

RUN cd worker-md && go mod download

COPY shared/ ./shared/
COPY worker-md/ ./worker-md/

RUN mkdir -p /out && \
    cd worker-md && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/worker .

# ── Stage 2: Run (needs wkhtmltopdf) ───────────────────────────
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y \
    wkhtmltopdf \
    libxrender1 \
    libxext6 \
    libfontconfig1 \
    ca-certificates \
    --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /out/worker .

EXPOSE 9090
CMD ["./worker"]
