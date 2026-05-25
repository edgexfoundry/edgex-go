package utils

import (
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/sirupsen/logrus"
	"io"
)

type LogrusAdaptor struct {
	logger logger.LoggingClient
}

func (f *LogrusAdaptor) Format(entry *logrus.Entry) ([]byte, error) {
	// Implement your custom formatting logic here
	return []byte(fmt.Sprintf("[%s] %s\n", entry.Level, entry.Message)), nil
}

func (f *LogrusAdaptor) Levels() []logrus.Level {
	return logrus.AllLevels
}

const OPENZITI_LOG_FORMAT = "openziti: %s"
const OPENZITI_DEFAULT_LOG_FORMAT = "default openziti: %s"

func (f *LogrusAdaptor) Fire(e *logrus.Entry) error {
	switch e.Level {
	case logrus.DebugLevel:
		f.logger.Debugf(OPENZITI_LOG_FORMAT, e.Message)
	case logrus.InfoLevel:
		f.logger.Infof(OPENZITI_LOG_FORMAT, e.Message)
	case logrus.WarnLevel:
		f.logger.Warnf(OPENZITI_LOG_FORMAT, e.Message)
	case logrus.ErrorLevel:
		f.logger.Errorf(OPENZITI_LOG_FORMAT, e.Message)
	case logrus.FatalLevel:
		f.logger.Errorf(OPENZITI_LOG_FORMAT, e.Message)
	case logrus.PanicLevel:
		f.logger.Errorf(OPENZITI_LOG_FORMAT, e.Message)
	default:
		f.logger.Errorf(OPENZITI_DEFAULT_LOG_FORMAT, e.Message)
	}

	return nil
}

func AdaptLogrusBasedLogging(lc logger.LoggingClient) {
	// Create a new logger instance
	hook := &LogrusAdaptor{
		logger: lc,
	}
	logrus.AddHook(hook)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
	})
	logrus.SetOutput(io.Discard)
}
