# Built from the REPO ROOT so the workspace-shared module is in context:
#   docker build -t docplatform-api:latest -f api/Dockerfile .
#
# ── Stage 1: Build ─────────────────────────────────────────────
FROM golang:1.25-alpine AS builder
ENV GOWORK=off
WORKDIR /src

# Module manifests first for layer caching. api/go.mod carries
# `replace github.com/docplatform/shared => ../shared`, so shared/go.mod
# must be present for dependency resolution.
COPY shared/go.mod shared/go.sum ./shared/
COPY api/go.mod api/go.sum ./api/
RUN cd api && go mod download

# Sources
COPY shared/ ./shared/
COPY api/ ./api/
RUN cd api && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/app .

# ── Stage 2: Run ───────────────────────────────────────────────
FROM alpine:3.19

# curl is used by the docker-compose healthcheck
RUN apk --no-cache add ca-certificates tzdata curl
WORKDIR /app
COPY --from=builder /out/app .

EXPOSE 8080
CMD ["./app"]
