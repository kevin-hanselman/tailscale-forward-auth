FROM golang:1.18-alpine AS build-env

WORKDIR /go/src/echo-server

COPY main.go go.mod ./
RUN go install .

FROM alpine
COPY --from=build-env /go/bin/* /usr/local/bin/
