# http3-server

A lightweight and minimal HTTP server implemented in Go for learning purposes, supporting HTTP/1.1 and HTTP/2 (HTTP/3 is abandoned).

This project is not a complete implementation and focuses on simplicity. A lot of features are missing.

Additionally, this is one of my first projects in Go, so there may be areas for improvement.

---

In its current state, the project meets my goals. I’ve learned the basics of Go, gained insights into how HTTP
protocols work (even without full implementation), and explored low-level programming concepts, such as bit manipulation.

The HTTP/2 version is far from implemented, unfortunately I don't have much time yet and I struggle with the stream/multiplexing part.
However, I already have learned a lot about how these protocols works under the hood, so this is it :)

Who knows, maybe in the future I will continue (and end) this project.

## Usage

*http/1.1*
```shell
go build && ./http3-server -p 8080 -v 1
curl -k https://localhost:8080

# the server responds with an index.html file and logs the following lines:
[12:28:32] HTTP1 server started to listen on port 8080

[12:28:40] New connection from [::1]:34728
[12:28:40] | GET /index.html HTTP/1.1
[12:28:40] | Host: localhost8080
[12:28:40] | User-Agent: curl/8.5.0
[12:28:40] | Accept: */*
[12:28:40] | <empty body>

[12:28:40] | HTTP/1.1 200 OK
[12:28:40] | Content-Type: text/html
[12:28:40] |
[12:28:40] | <body>
```

*http/2*
```shell
go build && ./http3-server -p 8080 -v 2
curl --http2-prior-knowledge -k https://localhost:8080

# the server logs the following lines:
[12:33:09] HTTP2 server started to listen on port 8080

[12:33:16] New connection from [::1]:40884
[12:33:16] "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
[12:33:16] +- SETTINGS
[12:33:16] | length: 18
[12:33:16] | flags: 
[12:33:16] | streamIdentifier: 0
[12:33:16] | payload: {
SETTINGS_MAX_CONCURRENT_STREAMS=100
SETTINGS_INITIAL_WINDOW_SIZE=10485760
SETTINGS_ENABLE_PUSH=0
}
[12:33:16] |
[12:33:16] Sent settings frame with ACK flag
```