# STAGE 1
FROM golang:1.26.5-alpine AS builder

WORKDIR /usr/src/app

COPY shared/go.mod shared/go.sum ./shared/
COPY admin/go.mod admin/go.sum ./admin/
RUN go work init ./admin
RUN go work edit -replace=sen1or/letslive/shared=./shared

WORKDIR /usr/src/app/admin
RUN go mod download

WORKDIR /usr/src/app
COPY shared/ ./shared/
COPY admin/ ./admin/

WORKDIR /usr/src/app/admin
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o /app ./cmd/

# STAGE 2
FROM alpine:latest AS final

WORKDIR /usr/src/app
COPY --from=builder /usr/src/app/admin/migrations /usr/src/app/migrations

COPY --from=builder /app /usr/local/bin/app

CMD ["app"]
