# STAGE 1
FROM golang:1.26.2-alpine AS builder

WORKDIR /usr/src/app

COPY shared/go.mod shared/go.sum ./shared/
COPY auth/go.mod auth/go.sum ./auth/
RUN go work init ./auth
RUN go work edit -replace=sen1or/letslive/shared=./shared

WORKDIR /usr/src/app/auth
RUN go mod download

WORKDIR /usr/src/app
COPY shared/ ./shared/
COPY auth/ ./auth/

WORKDIR /usr/src/app/auth
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o /app ./cmd/

# STAGE 2
FROM alpine:latest AS final

WORKDIR /usr/src/app
COPY --from=builder /usr/src/app/auth/migrations /usr/src/app/migrations

# copy the compiled binary executable
COPY --from=builder /app /usr/local/bin/app

CMD ["app"]
