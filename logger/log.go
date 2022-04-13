package logger

import (
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type Logger struct {
	zlog zerolog.Logger
}

func (logger *Logger) Infof(format string, args ...interface{}) {
	logger.zlog.Info().Msgf(format, args...)
}

func (logger *Logger) Debugf(format string, args ...interface{}) {
	logger.zlog.Debug().Msgf(format, args...)
}

func (logger *Logger) Errorf(format string, args ...interface{}) {
	logger.zlog.Error().Msgf(format, args...)
}

func (logger *Logger) Configure() {
	level, err := zerolog.ParseLevel(viper.GetString("verbosity"))

	if err != nil {
		logger.zlog.Fatal().Msgf("Failed to parse verbosity: %v", err)
	}

	logger.zlog = logger.zlog.Level(level)
}

func NewLogger() Logger {
	zlog := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	logger := Logger{zlog}
	logger.Configure()
	return logger
}
