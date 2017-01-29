FROM golang:latest

WORKDIR /go/src/github.com/ds0nt/nexus

COPY . ./

RUN go get