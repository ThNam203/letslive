# STAGE 1
FROM golang:1.26.5-alpine AS builder

WORKDIR /usr/src/app

COPY shared/go.mod shared/go.sum ./shared/
COPY realtime/go.mod realtime/go.sum ./realtime/
RUN go work init ./realtime
RUN go work edit -replace=sen1or/letslive/shared=./shared

WORKDIR /usr/src/app/realtime
RUN go mod download

WORKDIR /usr/src/app
COPY shared/ ./shared/
COPY realtime/ ./realtime/

WORKDIR /usr/src/app/realtime
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o /app ./cmd/

# STAGE 2
FROM alpine:latest AS final

COPY --from=builder /app /usr/local/bin/app

CMD ["app"]
