package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func NewLogger(component string) *zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}
	logger := zerolog.New(output).With().Str("component", component).Timestamp().Logger()

	return &logger
}
