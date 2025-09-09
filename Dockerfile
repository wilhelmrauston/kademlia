FROM golang:1.25-alpine

RUN apk update && apk add curl
WORKDIR /app


COPY go.mod ./
#COPY go.sum ./
RUN go mod download