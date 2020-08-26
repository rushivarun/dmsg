FROM golang:alpine as build-env

ENV GO111MODULE=on

RUN apk update && apk add bash ca-certificates git gcc g++ libc-dev

RUN mkdir /p2p
RUN mkdir -p /p2p/proto 

WORKDIR /p2p

COPY ./proto/service.pb.go /p2p/proto
COPY ./main.go /p2p

COPY go.mod .
COPY go.sum .

RUN go env

RUN go mod download

RUN go build -o p2p .

CMD ./p2p