FROM golang:1.18.3-alpine as builder

WORKDIR /app

COPY . /app
RUN go mod tidy
RUN go build -o /app main.go

FROM alpine:3
WORKDIR /app

RUN mkdir -p /var/local_storage/media/image/

COPY config.yaml .
COPY --from=builder /app/main /app
