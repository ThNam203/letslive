FROM golang:1.23

WORKDIR /usr/src/app

RUN apt-get update && apt-get install -y ffmpeg curl

COPY ../go.mod ../go.sum ./
RUN go mod download && go mod verify

COPY ../pkg ./pkg
COPY ../transcode ./transcode

RUN go build -v -o /usr/local/bin/app ./transcode/cmd/

CMD ["app"]
