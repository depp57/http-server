package tls

import (
	"crypto/tls"
	"errors"
	"http3-server/logger"
	"net"
)

func DecryptConnection(connection net.Conn) (net.Conn, error) {
	tlsConnection := tls.Server(connection, loadTlsConfig())
	err := tlsConnection.Handshake()
	if isPlainHttpConnection(err) {
		return nil, errors.New("unsupported plain http connection")
	}

	return tlsConnection, nil
}

func isPlainHttpConnection(err error) bool {
	switch err.(type) {
	case tls.RecordHeaderError:
		logger.Blue("Plain http connection detected")
		return true
	default:
		// other TLS errors, ignore since we use local certificate that causes validation errors
		return false
	}
}

func loadTlsConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair("tls/test.pem", "tls/test.key")
	if err != nil {
		logger.Fatal("Failed to load TLS cert/key")
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
}
