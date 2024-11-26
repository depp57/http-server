package http2

import (
	"cmp"
	"encoding/binary"
	"errors"
	"fmt"
	"http3-server/logger"
	"http3-server/tls"
	"net"
	"time"
)

const (
	crlf = "\r\n"

	frameHeaderLength = 9

	dataFrame         = 0x0
	headersFrame      = 0x1
	priorityFrame     = 0x2
	rstStreamFrame    = 0x3
	settingsFrame     = 0x4
	pushPromiseFrame  = 0x5
	pingFrame         = 0x6
	goAwayFrame       = 0x7
	windowUpdateFrame = 0x8
	continuationFrame = 0x9
)

var frameTypes = map[byte]string{
	dataFrame:         "DATA",
	headersFrame:      "HEADERS",
	priorityFrame:     "PRIORITY",
	rstStreamFrame:    "RST_STREAM",
	settingsFrame:     "SETTINGS",
	pushPromiseFrame:  "PUSH_PROMISE",
	pingFrame:         "PING",
	goAwayFrame:       "GOAWAY",
	windowUpdateFrame: "WINDOW_UPDATE",
	continuationFrame: "CONTINUATION",
}

var frameFlags = map[byte]map[byte]string{
	dataFrame: {
		0x1: "END_STREAM",
		0x8: "PADDED",
	},
	headersFrame: {
		0x1:  "END_STREAM",
		0x4:  "END_HEADERS",
		0x8:  "PADDED",
		0x20: "PRIORITY",
	},
	priorityFrame:  nil, // PRIORITY frame does not define any flags.
	rstStreamFrame: nil, // RST_STREAM frame does not define any flags.
	settingsFrame: {
		0x1: "ACK",
	},
	pushPromiseFrame: {
		0x4: "END_HEADERS",
		0x8: "PADDED",
	},
	pingFrame: {
		0x1: "ACK",
	},
	goAwayFrame:       nil, // GOAWAY frame does not define any flags.
	windowUpdateFrame: nil, // WINDOW_UPDATE frame does not define any flags.
	continuationFrame: {
		0x4: "END_HEADERS",
	},
}

var settingsFrameParameters = map[byte]string{
	0x1: "SETTINGS_HEADER_TABLE_SIZE",
	0x2: "SETTINGS_ENABLE_PUSH",
	0x3: "SETTINGS_MAX_CONCURRENT_STREAMS",
	0x4: "SETTINGS_INITIAL_WINDOW_SIZE",
	0x5: "SETTINGS_MAX_FRAME_SIZE",
	0x6: "SETTINGS_MAX_HEADER_LIST_SIZE",
}

type Server struct {
	listener net.Listener
}

type httpFrame struct {
	length           uint32
	frameType        byte
	flags            byte
	streamIdentifier uint32
	payload          string

	writeBuf []byte
}

func (frame *httpFrame) display() {
	frameType := frameTypes[frame.frameType]
	flags := frameFlags[frame.frameType][frame.flags]

	logger.Blue("+- %s", frameType)
	logger.Blue("| length: %d", frame.length)
	logger.Blue("| flags: %s", flags)
	logger.Blue("| streamIdentifier: %d", frame.streamIdentifier)
	logger.Blue("| payload: {%s}", frame.payload)
	logger.Blue("|")
}

func NewHttp2Server() *Server {
	return &Server{}
}

func (server *Server) ListenAndServe(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatal("Failed to start http server on port %d: %v", port, err)
	}

	logger.Gray("HTTP2 server started to listen on port %d", port)

	server.listener = listener

	server.serve()
}

func (server *Server) serve() {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			logger.Warn("Failed to accept connection: %v", err)
			continue
		}

		logger.Gray("New connection from %s", conn.RemoteAddr().String())

		tlsConn, err := tls.DecryptConnection(conn)
		if err != nil {
			//redirectToHttps(conn, port)
			logger.Warn("Failed to decrypt connection: %v", err)
			continue
		}

		tlsConn.SetDeadline(time.Now().Add(time.Second * 2))

		//upgradeHttpVersion(conn)

		go handleConnection(tlsConn)
	}
}

func handleConnection(conn net.Conn) {
	readConnectionPreface(conn)
	readFrame(conn)
}

// As per section 3.5
// https://httpwg.org/specs/rfc7540.html#rfc.section.3.5
// The client connection preface starts with a sequence of 24 octets
func readConnectionPreface(conn net.Conn) {
	bytes, err := readBytes(conn, 24)
	if err != nil {
		logger.Fatal("Failed to read connection preface: %v", err)
		return
	}

	logger.Blue("%q%s", bytes, crlf)
}

// As per section 4.
// https://httpwg.org/specs/rfc7540.html#rfc.section.4
// Once the HTTP/2 connection is established, endpoints can begin exchanging frames.
//
// All frames begin with a fixed 9-octet header followed by a variable-length payload.
//
// +-----------------------------------------------+
// |                 Length (24)                   |
// +---------------+---------------+---------------+
// |   Type (8)    |   Flags (8)   |
// +-+-------------+---------------+-------------------------------+
// |R|                 Stream Identifier (31)                      |
// +=+=============================================================+
// |                   Frame Payload (0...n)                      ...
// +---------------------------------------------------------------+
func readFrame(conn net.Conn) {
	frame := &httpFrame{}

	err := readFrameHeader(conn, frame)
	switch frame.frameType {
	case settingsFrame:
		readSettingsFramePayload(conn, frame)
		frame.display()
		respondSettingsFrame(conn, frame)
	}

	if err != nil {
		logger.Warn("Failed to read frame: %v", err)
		return
	}

}

func respondSettingsFrame(conn net.Conn, frame *httpFrame) {
	frame.startWrite(0x1) // ACK
	frame.endWrite()

	_, err := conn.Write(frame.writeBuf)
	if err != nil {
		logger.Warn("Failed to write settings frame: %v", err)
	} else {
		logger.Green("Sent settings frame with ACK flag")
	}
}

func (frame *httpFrame) startWrite(flags byte) {
	// write the header
	frame.writeBuf = append(frame.writeBuf[:0],
		0, // 3 bytes representing the length, filled later in endWrite
		0,
		0,
		frame.frameType,
		flags)
	frame.writeBuf = append(frame.writeBuf, encodeStreamIdentifier(frame.streamIdentifier)...)
}

func (frame *httpFrame) endWrite() {
	payloadLength := len(frame.writeBuf) - frameHeaderLength

	frame.writeBuf[0] = byte(payloadLength >> 16)
	frame.writeBuf[1] = byte(payloadLength >> 8)
	frame.writeBuf[2] = byte(payloadLength)
}

// Reads the 9-octet header
func readFrameHeader(conn net.Conn, frame *httpFrame) error {
	bytes, err := readBytes(conn, frameHeaderLength)
	if err != nil {
		return errors.New("failed to read the frame header")
	}

	frame.length = uint24ToUint32(bytes[:3])
	frame.frameType = bytes[3]
	frame.flags = bytes[4]
	frame.streamIdentifier = parseStreamIdentifier(bytes[5:])

	return nil
}

// As per section 6.5.1
// https://httpwg.org/specs/rfc7540.html#rfc.section.6.5.1
// The payload of a SETTINGS frame consists of zero or more parameters,
// each consisting of an unsigned 16-bit setting identifier and an unsigned 32-bit value.
//
// +-------------------------------+
// |       Identifier (16)         |
// +-------------------------------+-------------------------------+
// |                        Value (32)                             |
// +---------------------------------------------------------------+
func readSettingsFramePayload(conn net.Conn, frame *httpFrame) {
	bytes, err := readBytes(conn, int(frame.length))
	if err != nil {
		logger.Warn("Failed to read settings frame: %v", err)
	}

	for i := 0; i < len(bytes); i += 6 {
		identifier := bytes[i : i+2]
		parameter := settingsFrameParameters[identifier[1]]

		value := binary.BigEndian.Uint32(bytes[i+2 : i+6])

		frame.payload += fmt.Sprintf("%s%s=%d", crlf, parameter, value)
	}

	frame.payload += crlf
}

// Section 3.2 of HTTP/2 specification
// https://httpwg.org/specs/rfc7540.html#rfc.section.3.2
// A client that makes a request to an "http" URI without prior knowledge about support for HTTP/2
// uses the HTTP Upgrade mechanism.
//
// <-- GET / HTTP/1.1
// <-- Connection: Upgrade, HTTP2-Settings
// <-- Upgrade: h2
//
// --> HTTP/1.1 101 Switching Protocols
// --> Connection: Upgrade
// --> Upgrade: h2c
//
// Alternatively, a client can learn that a particular server supports HTTP/2 by other means.
// This way, it will directly send HTTP/2 requests.
func upgradeHttpVersion(conn net.Conn) {
	defer conn.Close()

	_, err1 := conn.Write([]byte(fmt.Sprintf("HTTP/1.1 101 Switching Protocols%s", crlf)))
	_, err2 := conn.Write([]byte(fmt.Sprintf("Connection: Upgrade%s", crlf)))
	_, err3 := conn.Write([]byte(fmt.Sprintf("Upgrade: h2%s", crlf)))
	_, err4 := conn.Write([]byte(crlf))

	if err := cmp.Or(err1, err2, err3, err4); err != nil {
		logger.Warn("Failed to write the response headers that upgrades the connection to http2: %v", err)
		return
	}
}
