# Built from the REPO ROOT so the workspace-shared module is in context:
#   docker build -t docplatform-worker-thumb:latest -f worker-thumb/Dockerfile .
#
# ── Stage 1: Build ─────────────────────────────────────────────
FROM golang:1.25-alpine AS builder
ENV GOWORK=off
WORKDIR /src

# worker-thumb/go.mod carries `replace github.com/docplatform/shared => ../shared`,
# so shared/go.mod must be present for dependency resolution.
COPY shared/go.mod shared/go.sum ./shared/
COPY worker-thumb/go.mod worker-thumb/go.sum ./worker-thumb/
RUN cd worker-thumb && go mod download

COPY shared/ ./shared/
COPY worker-thumb/ ./worker-thumb/
RUN cd worker-thumb && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/worker .

# ── Stage 2: Run (needs pdftoppm from poppler-utils) ───────────
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y \
    poppler-utils \
    ca-certificates \
    --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /out/worker .

EXPOSE 9090
CMD ["./worker"]
