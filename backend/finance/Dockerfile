# STAGE 1
FROM golang:1.26.2-alpine AS builder

WORKDIR /usr/src/app

COPY shared/go.mod shared/go.sum ./shared/
COPY finance/go.mod finance/go.sum ./finance/
RUN go work init ./finance
RUN go work edit -replace=sen1or/letslive/shared=./shared

WORKDIR /usr/src/app/finance
RUN go mod download

WORKDIR /usr/src/app
COPY shared/ ./shared/
COPY finance/ ./finance/

WORKDIR /usr/src/app/finance
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o /app ./cmd/

# STAGE 2
FROM alpine:latest AS final

WORKDIR /usr/src/app
COPY --from=builder /usr/src/app/finance/migrations /usr/src/app/migrations

COPY --from=builder /app /usr/local/bin/app

CMD ["app"]
