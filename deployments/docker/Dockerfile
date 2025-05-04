# argument for Go version
ARG GO_VERSION=1.22

# Use the official Golang image from DockerHub
FROM golang:${GO_VERSION}-alpine

RUN apk add git \
    && go install github.com/cespare/reflex@v0.3.1
WORKDIR /src

EXPOSE 8080