# App that has net/http package to run in alpine linux
# go build -ldflags '-linkmode external -w -extldflags "-static"'

FROM golang:1.14.6-buster AS builder

WORKDIR /go/src

ENV GO111MODULE=on
ENV GOFLAGS -mod=vendor

COPY ./ /go/src/

RUN GOOS=linux go build -v -ldflags '-linkmode external -w -extldflags "-static"' -o vault-gopher .

FROM debian:stable-20200607-slim

LABEL maintainer=roweluchi30@gmail.com

RUN apt-get install -y curl \
    && rm -rf /var/cache/apt/*

COPY --from=builder /go/src/vault-gopher /app/

WORKDIR /app

ENTRYPOINT [ "/app/vault-gopher" ]