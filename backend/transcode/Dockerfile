# STAGE 1
FROM golang:1.26.1 AS builder

WORKDIR /usr/src/app

COPY shared/go.mod shared/go.sum ./shared/
COPY transcode/go.mod transcode/go.sum ./transcode/
RUN go work init ./transcode
RUN go work edit -replace=sen1or/letslive/shared=./shared

WORKDIR /usr/src/app/transcode
RUN go mod download

WORKDIR /usr/src/app
COPY shared/ ./shared/
COPY transcode/ ./transcode/

WORKDIR /usr/src/app/transcode
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o /app ./cmd/

# STAGE 2
FROM alpine:latest AS final

RUN apk update && \
    apk add --no-cache \
        ffmpeg \
        curl \
        ca-certificates \
        tzdata && \
    rm -rf /var/cache/apk/*

WORKDIR /usr/src/app

COPY --from=builder /app /usr/local/bin/app

CMD ["app"]
