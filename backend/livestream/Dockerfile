FROM golang:1.24.0

WORKDIR /usr/src/app

COPY ../go.mod ../go.sum ./

RUN go mod download 
RUN go mod verify

COPY ../ .

RUN go build -v -o /usr/local/bin/app ./cmd/

CMD ["app"]
