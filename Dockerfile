FROM golang:1.18.3-alpine

WORKDIR /app

RUN mkdir bin
RUN mkdir src
RUN mkdir -p /var/local_storage/media/image/

COPY go.mod .
COPY go.sum .

RUN go mod download
RUN go mod tidy

WORKDIR /app/src

COPY . .
RUN go build -o /app/bin/main main.go

RUN rm -r /app/src/

WORKDIR /app/bin

COPY config.yaml .

RUN mkdir db
COPY db db
