FROM golang:latest

MAINTAINER Brian Myers (brian.carl.myers@gmail.com)

ADD . /go/src/github.com/bcmyers/gogate

WORKDIR /go/src/github.com/bcmyers/gogate

RUN go get github.com/go-redis/redis github.com/Sirupsen/logrus github.com/sebest/logrusly
RUN go install

ENTRYPOINT /go/bin/gogate

EXPOSE 80
