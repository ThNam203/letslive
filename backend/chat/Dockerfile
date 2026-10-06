# STAGE 1
FROM golang:1.26.5-alpine AS builder

WORKDIR /usr/src/app

COPY shared/go.mod shared/go.sum ./shared/
COPY chat/go.mod chat/go.sum ./chat/
RUN go work init ./chat
RUN go work edit -replace=sen1or/letslive/shared=./shared

WORKDIR /usr/src/app/chat
RUN go mod download

WORKDIR /usr/src/app
COPY shared/ ./shared/
COPY chat/ ./chat/

WORKDIR /usr/src/app/chat
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o /app ./cmd/

# STAGE 2
FROM alpine:latest AS final

COPY --from=builder /app /usr/local/bin/app

CMD ["app"]
