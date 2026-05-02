# ── Stage 1: Build ─────────────────────────────────────────────
FROM golang:bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o worker .

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
COPY --from=builder /app/worker .

CMD ["./worker"]