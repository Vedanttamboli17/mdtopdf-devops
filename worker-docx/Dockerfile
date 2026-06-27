# Built from the REPO ROOT so the workspace-shared module is in context:
#   docker build -t docplatform-worker-docx:latest -f worker-docx/Dockerfile .
#
# ── Stage 1: Build ─────────────────────────────────────────────
FROM golang:1.25.3 AS builder
ENV GOWORK=off
WORKDIR /src

# worker-docx/go.mod carries `replace github.com/docplatform/shared => ../shared`,
# so shared/go.mod must be present for dependency resolution.
COPY shared/go.mod shared/go.sum ./shared/
COPY worker-docx/go.mod worker-docx/go.sum ./worker-docx/
RUN cd worker-docx && go mod download

COPY shared/ ./shared/
COPY worker-docx/ ./worker-docx/
RUN cd worker-docx && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/worker .

# ── Stage 2: Run (needs LibreOffice) ───────────────────────────
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y \
    libreoffice-writer \
    libreoffice-calc \
    libreoffice-impress \
    fonts-liberation \
    fonts-dejavu \
    ca-certificates \
    --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

# LibreOffice headless requires a writable HOME
ENV HOME=/root

WORKDIR /app
COPY --from=builder /out/worker .

EXPOSE 9090
CMD ["./worker"]
