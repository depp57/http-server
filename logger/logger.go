package logger

import (
	"fmt"
	"os"
	"time"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	green  = "\033[32m"
	white  = "\033[37m"
)

func Gray(format string, v ...any) {
	log(white, format, v...)
}

func Blue(format string, v ...any) {
	log(blue, format, v...)
}

func Green(format string, v ...any) {
	log(green, format, v...)
}

func Warn(format string, v ...any) {
	log(yellow, format, v...)
}

func Fatal(format string, v ...any) {
	log(red, format, v...)
	os.Exit(1)
}

func log(color string, format string, v ...any) {
	now := time.Now().Format("15:04:05")

	printV := []any{color, now}
	printV = append(printV, v...)
	printV = append(printV, reset)

	fmt.Printf("%s[%s] "+format+"%s\n", printV...)
}
