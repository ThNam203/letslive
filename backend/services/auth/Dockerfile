FROM golang:1.22

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify && go mod tidy

COPY . .
RUN go build -v -o /usr/local/bin/app ./cmd/

# the set-udp-buffer-size.sh doesn't work smh
# TODO: somehow make it works please
# CMD ["/bin/bash", "-c", "./set-udp-buffer-size.sh && /usr/local/bin/app"]
CMD ["app"]
