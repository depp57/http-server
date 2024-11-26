package servers

import (
	"http3-server/logger"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func SendFile(path string, conn net.Conn) {
	content, err := os.ReadFile(path)
	if err != nil {
		logger.Fatal("cannot read file '%s'", path)
	}
	if _, err = conn.Write(content); err != nil {
		logger.Warn("cannot write body to incoming connection")
	}
}

func GetFileExtension(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}
