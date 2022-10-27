package logger

import (
	"io"
	"log"
	"os"
)

func New(prefix string, outFile string) (*log.Logger, error) {
	var w io.Writer = log.Writer()
	if outFile != "" {
		flags := os.O_APPEND | os.O_CREATE | os.O_WRONLY
		file, err := os.OpenFile(outFile, flags, 0600) //#nosec G304
		if err != nil {
			return nil, err
		}
		w = io.MultiWriter(w, file)
	}

	logger := log.New(w, prefix, log.LstdFlags|log.Lshortfile)
	return logger, nil
}
