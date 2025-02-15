FROM golang:1.24.0

WORKDIR /usr/src/app

RUN apt-get update && apt-get install -y ffmpeg curl

COPY ../go.mod ../go.sum ./
RUN go mod download
RUN go mod verify

COPY ../ .

RUN go build -v -o /usr/local/bin/app ./transcode/cmd/

CMD ["app"]
