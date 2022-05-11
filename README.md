# tailscale-forward-auth

This is a basic example of how to implement a Tailscale authentication server
for general use with proxies. It is derived from the [Tailscale nginx-auth
command](https://github.com/tailscale/tailscale/blob/741ae9956e674177687062b5499a80db83505076/cmd/nginx-auth/README.md),
but it is decoupled from NGINX and packaged in a Docker image.

## Traefik example

**Don't run this example in production; it's not secure.**

The `example` directory contains an example of running the server as a
[ForwardAuth middleware in
Traefik](https://doc.traefik.io/traefik/middlewares/http/forwardauth/). It
assumes that you have Docker and Docker Compose available on your machine, and
that tailscaled is running and authenticated on the same machine (using the
tailscaled UNIX socket at the default location).

To run the example:

    cd example
    docker-compose up -d
    docker-compose logs -f

From the same machine, send an HTTP request to the proxied application using
its local IP address. You should receive a 401 error:

```
$ curl -v localhost/echo
*   Trying 127.0.0.1:80...
* Connected to localhost (127.0.0.1) port 80 (#0)
> GET /echo HTTP/1.1
> Host: localhost
> User-Agent: curl/7.83.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 401 Unauthorized
< Content-Length: 0
< Date: Sat, 07 May 2022 18:44:39 GMT
<
* Connection #0 to host localhost left intact
```

Now send an HTTP request using the Tailscale IP. You should now receive
a response from your echo server. Note the added `Tailscale-` fields:

```
$ curl $(tailscale ip -4)/echo
{
  "Accept": [
    "*/*"
  ],
  "Accept-Encoding": [
    "gzip"
  ],
  "Tailscale-Login": [
    "jane-doe"
  ],
  "Tailscale-Name": [
    "Jane Doe"
  ],
  "Tailscale-Profile-Picture": [],
  "Tailscale-Tailnet": [
    "jane-doe.example"
  ],
  "Tailscale-User": [
    "jane-doe@example"
  ],
  "User-Agent": [
    "curl"
  ],
  "X-Forwarded-For": [
    "100.x.x.x"
  ],
  "X-Forwarded-Host": [
    "100.x.x.x"
  ],
  "X-Forwarded-Port": [
    "80"
  ],
  "X-Forwarded-Proto": [
    "http"
  ],
  "X-Forwarded-Server": [
    "xxxxxxxxxx"
  ],
  "X-Real-Ip": [
    "100.x.x.x"
  ]
}
```
