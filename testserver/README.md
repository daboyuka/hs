# Simple HTTP test server

Usage: `go run . [-s] [bindaddr]`

Listens on `bindaddr` (default `:8080`) for HTTP requests (HTTPS if `-s` specified),
and logs the Go `http.Request` and returns status 200 on request.

## Using HTTPS
To listen with HTTPS, use `testserver/key.pem` to authenticate the server. This is a self-signed
certificate, so some clients may reject it for that reason.

Example:
```
# Note: -k for insecure, because the cert is self-signed
$ curl -k --capath testserver 'https://localhost:80/some/path'
```
