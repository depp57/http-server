package servers

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func readFileInto(path string, writer io.Writer) {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	writer.Write(content)
}

func getFileExtension(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}
