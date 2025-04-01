# STAGE 1
FROM golang:1.24.1 AS builder

WORKDIR /usr/src/app

COPY ./go.mod ./go.sum .
RUN go mod download
RUN go mod verify

COPY . .

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
