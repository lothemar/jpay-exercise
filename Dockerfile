# syntax=docker/dockerfile:1

FROM golang:1.17-bullseye

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY sample.db ./
RUN go build -o main

EXPOSE 8080

CMD [ "/app/main" ]
