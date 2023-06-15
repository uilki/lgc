package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

type ServerLogger struct {
	*logrus.Logger
}

type Logger interface {
	LogError(e interface{})
	LogInfo(info interface{})
	LogDebug(message string, arg ...any)
}

func NewLogger() *ServerLogger {
	logger := &ServerLogger{logrus.New()}

	logger.SetOutput(os.Stdout)

	logger.Formatter = &logrus.TextFormatter{
		ForceColors:            true,
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		PadLevelText:           true,
		TimestampFormat:        "02/Jan/2006:15:04:05 -0700",
	}
	return logger
}

func (sl *ServerLogger) LogError(e interface{}) {
	sl.Errorln(e)
}

func (sl *ServerLogger) LogInfo(info interface{}) {
	sl.Infoln(info)
}

func (sl *ServerLogger) LogDebug(message string, arg ...any) {
	sl.Debugf(message, arg...)
}
