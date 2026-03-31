# STAGE 1
FROM golang:1.26.0-alpine AS builder

WORKDIR /usr/src/app

COPY shared/go.mod shared/go.sum ./shared/
COPY vod/go.mod vod/go.sum ./vod/
RUN go work init ./vod
RUN go work edit -replace=sen1or/letslive/shared=./shared

WORKDIR /usr/src/app/vod
RUN go mod download

WORKDIR /usr/src/app
COPY shared/ ./shared/
COPY vod/ ./vod/

WORKDIR /usr/src/app/vod
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o /app ./cmd/

# STAGE 2
FROM alpine:latest AS final

WORKDIR /usr/src/app
COPY --from=builder /usr/src/app/vod/migrations /usr/src/app/migrations

# Copy the compiled binary from the builder stage
COPY --from=builder /app /usr/local/bin/app

CMD ["app"]
