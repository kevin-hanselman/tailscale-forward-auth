FROM golang:1.19.3-alpine AS build-env

WORKDIR /go/src/tailscale-forward-auth

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 go install .

FROM scratch
COPY --from=build-env /go/bin/tailscale-forward-auth /tailscale-forward-auth
ENTRYPOINT ["/tailscale-forward-auth"]
