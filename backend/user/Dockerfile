# STAGE 1
FROM golang:1.26.2-alpine AS builder

WORKDIR /usr/src/app

COPY shared/go.mod shared/go.sum ./shared/
COPY user/go.mod user/go.sum ./user/
RUN go work init ./user
RUN go work edit -replace=sen1or/letslive/shared=./shared

WORKDIR /usr/src/app/user
RUN go mod download

WORKDIR /usr/src/app
COPY shared/ ./shared/
COPY user/ ./user/

WORKDIR /usr/src/app/user
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o /app ./cmd/

# STAGE 2
FROM alpine:latest AS final

WORKDIR /usr/src/app
COPY --from=builder /usr/src/app/user/migrations /usr/src/app/migrations

COPY --from=builder /app /usr/local/bin/app

CMD ["app"]
