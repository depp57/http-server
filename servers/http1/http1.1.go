package http1

import (
	"bufio"
	"cmp"
	"fmt"
	"http3-server/logger"
	"http3-server/servers"
	"http3-server/tls"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	headerContentType   = "Content-Type"
	headerContentLength = "Content-Length"

	crlf = "\r\n"

	defaultHeadersCount = 10
)

type Http1Server struct {
	listener net.Listener
}

type Request struct {
	path     string
	method   string
	protocol string
	headers  map[string]string
	body     string
}

func NewHttp1Server() *Http1Server {
	return &Http1Server{}
}

func (server *Http1Server) ListenAndServe(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatal("Failed to start http server on port %d: %v", port, err)
	}

	logger.Gray("HTTP1 server started to listen on port %d", port)

	server.listener = listener

	server.serve(port)
}

func (server *Http1Server) serve(port int) {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			logger.Warn("Failed to accept connection: %v", err)
			continue
		}

		logger.Gray("New connection from %s", conn.RemoteAddr().String())

		tlsConn, err := tls.DecryptConnection(conn)
		if err != nil {
			redirectToHttps(conn, port)
			continue
		}

		tlsConn.SetDeadline(time.Now().Add(time.Second * 2))

		go handleConnection(tlsConn)
	}
}

func redirectToHttps(conn net.Conn, port int) {
	defer conn.Close()

	logger.Blue("redirecting to https")
	_, err1 := conn.Write([]byte("HTTP/1.1 301 Moved Permanently\r\n"))
	_, err2 := conn.Write([]byte(fmt.Sprintf("Location: https://localhost:%d%s", port, crlf)))
	_, err3 := conn.Write([]byte(crlf))

	if err := cmp.Or(err1, err2, err3); err != nil {
		logger.Warn("Failed to write the permanent redirection header")
	}
}

func (req *Request) display() {
	logger.Blue("| %s %s %s", req.method, req.path, req.protocol)

	for key, value := range req.headers {
		logger.Blue("| %s: %s", key, value)
	}

	if len(req.body) > 0 {
		logger.Blue("| <body>")
		logger.Blue("| %s", req.body)
	} else {
		logger.Blue("| <empty body>")
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	req, err := handleRequest(conn)
	if err != nil {
		logger.Warn("Failed to handle request: %v", err)
		return
	}

	req.display()
	respond(req, conn)
}

// Build a Request from a http/1.1 message
//
// Example of a message
// GET / HTTP/1.1
// Host: www.example.com
// User-Agent: Mozilla/5.0
// Accept: text/html
// Accept-Encoding: gzip, deflate, br
// Connection: keep-alive
//
// foo=bar
func handleRequest(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)

	method, path, protocol, err := readFirstLine(reader)
	if err != nil {
		return nil, err
	}
	request := &Request{
		method:   method,
		path:     resolvePath(path),
		protocol: protocol,
		headers:  make(map[string]string, defaultHeadersCount),
	}

	err = readHeaders(reader, request)
	if err != nil {
		return nil, err
	}

	readBody(reader, request)

	return request, nil
}

// Example of a first line
// GET / HTTP/1.1
func readFirstLine(reader *bufio.Reader) (string, string, string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		logger.Warn("Failed to the read the first line: %v", err)
		return "", "", "", err
	}

	line = strings.TrimSpace(line)
	parts := strings.Split(line, " ")

	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid first line: %s", line)
	}

	return parts[0], parts[1], parts[2], nil
}

func readHeaders(reader *bufio.Reader, req *Request) error {
	for {
		key, value, err := readNextHeader(reader)
		if err != nil {
			return err
		} else if key == "" {
			break
		}

		req.headers[key] = value
	}

	return nil
}

// Example of a header
// Host: www.example.com
//
// Or it could be an empty line, announcing that there is no more header
func readNextHeader(reader *bufio.Reader) (string, string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		logger.Warn("Failed to the read the header: %v", err)
		return "", "", err
	}

	line = strings.TrimSpace(line)
	parts := strings.Split(line, ":")

	// empty line: no more header
	if (len(parts) == 0) && parts[0] == "" {
		return "", "", nil
	}

	key := parts[0]
	value := strings.TrimSpace(strings.Join(parts[1:], ""))

	return key, value, nil
}

// Example of body
// foo=bar
// hello, world!
func readBody(reader *bufio.Reader, req *Request) {
	// as per section 8.6 of http semantic, A user agent SHOULD send Content-Length in a request. But MUST NOT.
	// https://www.rfc-editor.org/rfc/rfc9110#section-8.6
	contentLength, err := strconv.Atoi(req.headers[headerContentLength])
	if err != nil {
		contentLength = reader.Buffered()
	}

	body, err := reader.Peek(contentLength)
	if err != nil {
		logger.Warn("Failed to read the body: %v", err)
	}

	req.body = string(body)
}

// as per section 6 of HTTP/1.1 protocol: https://www.w3.org/Protocols/rfc2616/rfc2616-sec6.html
//
// response = Status-Line
//
//		      *(( general-header
//		       |  response-header
//		       |  entity-header ) CRLF)
//		      CRLF
//		      [ message-body ]
//
//	with Status-Line = HTTP-Version SP Status-Code SP Reason-Phrase CRLF
//	and *-header being different type of header
func respond(req *Request, conn net.Conn) {
	contentType := fmt.Sprintf("text/%s", servers.GetFileExtension(req.path))

	_, err1 := conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK%s", crlf)))
	_, err2 := conn.Write([]byte(fmt.Sprintf("%s: %s%s", headerContentType, contentType, crlf)))
	_, err3 := conn.Write([]byte(crlf))

	if err := cmp.Or(err1, err2, err3); err != nil {
		logger.Warn("Failed to write the response body: %v", err)
		return
	}

	servers.SendFile("public"+req.path, conn)

	logger.Green("| HTTP/1.1 200 OK")
	logger.Green("| %s: %s", headerContentType, contentType)
	logger.Green("|")
	logger.Green("| <body>")
}

func resolvePath(path string) string {
	if path == "/" {
		return path + "index.html"
	}

	return path
}
