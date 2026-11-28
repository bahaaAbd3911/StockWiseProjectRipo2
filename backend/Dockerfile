# Production image: small static binaries (api, migrate, platform) and the SQL
# migrations on a minimal base with no shell, running as a non-root user.
#
#   docker build -t stockwise-api .
#   docker run --env-file .env -p 8080:8080 stockwise-api            # the API
#   docker run --env-file .env stockwise-api /app/migrate up         # migrations

# ---- build ----
FROM golang:1.27.4-alpine AS build
WORKDIR /src

# Download modules first so this layer is cached until go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Static binaries (no C library needed), without debug info, with paths of
# this machine stripped out.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/platform ./cmd/platform \
 && mkdir -p /out/data/uploads /out/data/certs

# ---- run ----
# distroless/static: CA certificates and time zones only; no shell, no package
# manager. :nonroot runs as uid 65532.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/api /out/migrate /out/platform /app/
COPY --from=build /src/migrations /app/migrations
# Writable folders for uploaded images and Let's Encrypt certificates. Mount
# volumes here so they survive new versions of the container.
COPY --from=build --chown=nonroot:nonroot /out/data /app/data

ENV APP_PORT=8080 \
    MIGRATIONS_DIR=/app/migrations \
    UPLOAD_DIR=/app/data/uploads \
    TLS_CACHE_DIR=/app/data/certs
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/app/api", "healthcheck"]

ENTRYPOINT []
CMD ["/app/api"]
