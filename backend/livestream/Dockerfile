# STAGE 1
FROM golang:1.26.2-alpine AS builder

WORKDIR /usr/src/app

COPY shared/go.mod shared/go.sum ./shared/
COPY livestream/go.mod livestream/go.sum ./livestream/
RUN go work init ./livestream
RUN go work edit -replace=sen1or/letslive/shared=./shared

WORKDIR /usr/src/app/livestream
RUN go mod download

WORKDIR /usr/src/app
COPY shared/ ./shared/
COPY livestream/ ./livestream/

WORKDIR /usr/src/app/livestream
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o /app ./cmd/

# STAGE 2
FROM alpine:latest AS final

WORKDIR /usr/src/app
COPY --from=builder /usr/src/app/livestream/migrations /usr/src/app/migrations

# Copy the compiled binary from the builder stage
COPY --from=builder /app /usr/local/bin/app

CMD ["app"]
