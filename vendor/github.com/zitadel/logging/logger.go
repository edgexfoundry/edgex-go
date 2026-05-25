package logging

import (
	"io"

	"github.com/sirupsen/logrus"
)

type logger logrus.Logger

var log *logger = (*logger)(logrus.StandardLogger())

func SetOutput(out io.Writer) {
	(*logrus.Logger)(log).SetOutput(out)
}

func SetFormatter(formatter logrus.Formatter) {
	(*logrus.Logger)(log).SetFormatter(formatter)
}

func SetLevel(level logrus.Level) {
	(*logrus.Logger)(log).SetLevel(level)
}

func SetGlobal() {
	logrus.SetFormatter(log.Formatter)
	logrus.SetLevel(log.Level)
	logrus.SetReportCaller(log.ReportCaller)
	logrus.SetOutput(log.Out)
	log = (*logger)(logrus.StandardLogger())
}
