package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

type ServerLogger struct {
	*logrus.Logger
}

func NewLogger() *ServerLogger {
	logger := &ServerLogger{logrus.New()}
	f, err := os.OpenFile("chat-server-go.log", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0664)
	if err != nil {
		f = os.Stdout
	}

	logger.SetOutput(f)

	logger.Formatter = &logrus.TextFormatter{ForceColors: true, FullTimestamp: true, DisableLevelTruncation: true, PadLevelText: true}
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
