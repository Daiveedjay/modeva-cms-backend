# ══════════════════════════════════════════════════════════════
# Production image for the Modeva CMS backend.
#
# Local development does NOT use this file - `make dev` runs Air
# natively on Windows and docker-compose.yml only starts Postgres
# and Redis. Air is therefore deliberately absent here.
#
# Tool versions are pinned on purpose: an unpinned `air@latest`
# started requiring Go >= 1.26 and broke this build with no code
# change on our side.
# ══════════════════════════════════════════════════════════════

# ── Stage 1: build ────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

# migrate is a runtime dependency (start.sh runs migrations on boot),
# built here so the final image needs no Go toolchain.
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

WORKDIR /src

# Cached separately from the source so dependency downloads survive code edits.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO off: pgx and gorm's postgres driver are pure Go, so this produces a
# static binary that runs on a bare Alpine with no libc shims.
# -trimpath strips local paths, -s -w drops debug tables.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/modeva-api .

# ── Stage 2: runtime ──────────────────────────────────────────
# Matches the Alpine release the golang:1.25-alpine image is built on.
FROM alpine:3.24

# bash             - start.sh is a bash script, not POSIX sh
# postgresql-client - start.sh uses psql for schema and extension setup
# ca-certificates  - outbound TLS to Postgres, Cloudinary, Resend, Google OAuth
RUN apk add --no-cache bash postgresql-client ca-certificates

WORKDIR /app

COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /out/modeva-api /app/modeva-api
COPY migrations ./migrations
COPY start.sh ./start.sh
RUN chmod +x ./start.sh

# Drop root: the process only reads from /app and opens a socket.
RUN addgroup -S modeva && adduser -S -G modeva modeva
USER modeva

EXPOSE 8081

CMD ["/app/start.sh"]
