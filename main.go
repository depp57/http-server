package main

import (
	"flag"
	"fmt"
	"http3-server/servers/http1"
	"http3-server/servers/http2"
	"http3-server/servers/http3"
)

type httpServer interface {
	ListenAndServe(port int)
}

func main() {
	var version, port int

	flag.IntVar(&version, "v", 1, "shorthand for version.")
	flag.IntVar(&version, "version", 1, "specify the http version of the server. Must be either 1, 2 or 3.")

	flag.IntVar(&port, "p", 8080, "shorthand for port.")
	flag.IntVar(&port, "port", 8080, "specify the port the server will listen on.")

	flag.Parse()

	if flag.NFlag() == 0 {
		fmt.Println("No arguments provided. Running the server with default http version (1) and port (8080).")
		fmt.Println("Run with -h to display the usage")
	}

	if flag.NFlag() > 2 {
		flag.Usage()
	}

	var server httpServer

	switch version {
	case 1:
		server = http1.NewHttp1Server()
	case 2:
		server = http2.NewHttp2Server()
	case 3:
		server = http3.NewHttp3Server()
	default:
		flag.Usage()
	}

	server.ListenAndServe(port)
}
