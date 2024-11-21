package main

import (
	"flag"
	"fmt"
	"http3-server/servers"
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

	if flag.NFlag() > 1 {
		flag.Usage()
	}

	var server httpServer

	switch version {
	case 1:
		server = servers.NewHttp1Server()
	case 2:
		server = servers.NewHttp2Server()
	case 3:
		server = servers.NewHttp3Server()
	default:
		panic("unknown version")
	}

	server.ListenAndServe(port)
}
