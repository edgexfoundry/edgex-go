package pfxlog

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type formatter struct {
	options *Options
}

func NewFormatter(options *Options) logrus.Formatter {
	return &formatter{options}
}

func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	var timeLabel string
	if f.options.AbsoluteTime {
		timeLabel = "[" + time.Now().Format(f.options.PrettyTimestampFormat) + "]"
	} else {
		seconds := time.Since(f.options.StartTimestamp).Seconds()
		timeLabel = fmt.Sprintf("[%8.3f]", seconds)
	}
	var level string
	switch entry.Level {
	case logrus.PanicLevel:
		level = f.options.PanicLabel
	case logrus.FatalLevel:
		level = f.options.FatalLabel
	case logrus.ErrorLevel:
		level = f.options.ErrorLabel
	case logrus.WarnLevel:
		level = f.options.WarningLabel
	case logrus.InfoLevel:
		level = f.options.InfoLabel
	case logrus.DebugLevel:
		level = f.options.DebugLabel
	case logrus.TraceLevel:
		level = f.options.TraceLabel
	}
	trimmedFunction := ""
	if entry.Caller != nil {
		trimmedFunction = strings.TrimPrefix(entry.Caller.Function, f.options.TrimPrefix)
	} else if fn, found := entry.Data["func"]; found {
		if fns, ok := fn.(string); ok {
			trimmedFunction = strings.TrimPrefix(fns, f.options.TrimPrefix)
			delete(entry.Data, "func")
			delete(entry.Data, "file")
		}
	}
	if v, found := entry.Data["_channels"]; found {
		trimmedFunction += " |"
		for i, channel := range v.([]string) {
			if i > 0 {
				trimmedFunction += ", "
			}
			trimmedFunction += channel
		}
		trimmedFunction += "|"
	}
	if context, found := entry.Data["_context"]; found {
		trimmedFunction += " [" + context.(string) + "]"
	}
	message := entry.Message
	if withFields(entry.Data) {
		fields := "{"
		field := 0
		for k, v := range entry.Data {
			if k != "_context" && k != "_channels" {
				if field > 0 {
					fields += " "
				}
				field++
				fields += fmt.Sprintf("%s=[%v]", k, v)
			}
		}
		fields += "} "
		message = f.options.FieldsColor + fields + f.options.DefaultFgColor + message
	}
	return []byte(fmt.Sprintf("%s %s %s: %s\n",
			f.options.TimestampColor+timeLabel+f.options.DefaultFgColor,
			level,
			f.options.FunctionColor+trimmedFunction+f.options.DefaultFgColor,
			message),
		),
		nil
}

func withFields(data map[string]interface{}) bool {
	gt := 0
	if _, contextFound := data["_context"]; contextFound {
		gt++
	}
	if _, channelsFound := data["_channels"]; channelsFound {
		gt++
	}
	return len(data) > gt
}
