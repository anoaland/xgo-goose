package xgoose

import "github.com/rs/zerolog"

type ZeroLogGooseLogger struct {
	logger *zerolog.Logger
}

func (l *ZeroLogGooseLogger) Printf(format string, v ...interface{}) {
	l.logger.Info().Msgf(format, v...)
}

func (l *ZeroLogGooseLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatal().Msgf(format, v...)
}

func NewZeroLogGooseLogger(logger *zerolog.Logger) *ZeroLogGooseLogger {
	return &ZeroLogGooseLogger{logger}
}
