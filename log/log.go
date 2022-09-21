package log

import (
	"io"
	"os"

	mlog "github.com/google/martian/v3/log"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var Logger = NewLogger()

func NewLogger() logger {
	return logger{Zlog: zerolog.New(os.Stderr).With().Timestamp().Logger()}
}

type logger struct {
	Zlog zerolog.Logger
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.Zlog.Info().Msgf(format, args...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.Zlog.Debug().Msgf(format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.Zlog.Error().Msgf(format, args...)
}

func (l *logger) Configure() {
	level, err := zerolog.ParseLevel(viper.GetString("verbosity"))

	if err != nil {
		l.Zlog.Fatal().Msgf("Failed to parse verbosity: %v", err)
	}

	var output io.Writer

	if viper.GetBool("pretty") {
		output = zerolog.NewConsoleWriter()
	} else {
		output = os.Stderr
	}

	l.Zlog = l.Zlog.Level(level).Output(output)
	mlog.SetLogger(l)
}

func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

func Configure() {
	Logger.Configure()
}
