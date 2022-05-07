FROM golang:1.18-alpine AS build-env

WORKDIR /go/src/tailscale-forward-auth

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN go install .

FROM alpine
COPY --from=build-env /go/bin/* /usr/local/bin/
