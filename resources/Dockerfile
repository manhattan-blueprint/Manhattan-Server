FROM golang:1.11.1-alpine3.8

WORKDIR /go/src/github.com/jaylees14/Manhattan-Server/

COPY . .

RUN apk add --no-cache git &&\
    go get github.com/gorilla/mux &&\
    go build -o resources

ENTRYPOINT ./resources

EXPOSE 80