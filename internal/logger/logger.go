package logger

import (
	"os"

	"github.com/rs/zerolog"
)

var log *zerolog.Logger = nil

func NewLogger() {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}
	logger := zerolog.New(output).With().Timestamp().Logger()

	log = &logger
}

func Infof(format string, args ...interface{}) {
	if log == nil {
		NewLogger()
	}

	log.Info().Msgf(format, args...)
}
