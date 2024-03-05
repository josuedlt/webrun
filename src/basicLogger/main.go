package basicLogger

import (
	"io"
	"log"
	"os"
)

func CreateLogger(silent bool, logFile string) *log.Logger {
	output := []io.Writer{}
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			output = append(output, file)
		}
	}
	if !silent {
		output = append(output, os.Stderr)
	}
	return log.New(io.MultiWriter(output...), "", log.LstdFlags)
}
