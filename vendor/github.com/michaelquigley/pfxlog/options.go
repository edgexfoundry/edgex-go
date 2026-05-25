package pfxlog

import (
	"fmt"
	"github.com/mgutz/ansi"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	StartTimestamp time.Time
	AbsoluteTime   bool
	TrimPrefix     string

	PanicLabel   string
	FatalLabel   string
	ErrorLabel   string
	WarningLabel string
	InfoLabel    string
	DebugLabel   string
	TraceLabel   string

	TimestampColor string
	FunctionColor  string
	FieldsColor    string
	DefaultFgColor string

	PrettyTimestampFormat string
	JsonTimestampFormat   string

	ActiveChannels           map[string]struct{}
	ChannelLogLevelOverrides map[string]logrus.Level

	DataFielder    func(data interface{}, entry *logrus.Entry) *logrus.Entry
	EnabledChecker func(data interface{}) bool

	StandardLogger *logrus.Logger
	Loggers        map[logrus.Level]*logrus.Logger
}

func DefaultOptions() *Options {
	options := &Options{
		StartTimestamp:        time.Now(),
		AbsoluteTime:          false,
		PrettyTimestampFormat: "2006-01-02 15:04:05.000",
		JsonTimestampFormat:   "2006-01-02T15:04:05.000Z",
		StandardLogger:        logrus.StandardLogger(),
		Loggers:               map[logrus.Level]*logrus.Logger{},
	}

	if defaultEnv("PFXLOG_USE_COLOR", false) {
		return options.Color()
	} else {
		return options.NoColor()
	}
}

func (options *Options) Starting(t time.Time) *Options {
	options.StartTimestamp = t
	return options
}

func (options *Options) StartingToday() *Options {
	now := time.Now()
	options.StartTimestamp = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return options
}

func (options *Options) SetAbsoluteTime() *Options {
	options.AbsoluteTime = true
	return options
}

func (options *Options) SetTrimPrefix(prefix string) *Options {
	options.TrimPrefix = prefix
	return options
}

func (options *Options) SetActiveChannels(channels ...string) *Options {
	for _, channel := range channels {
		if options.ActiveChannels == nil {
			options.ActiveChannels = make(map[string]struct{})
		}
		options.ActiveChannels[channel] = struct{}{}
	}
	return options
}

func (options *Options) SetChannelLogLevel(channel string, level logrus.Level) {
	m := map[string]logrus.Level{}
	for k, v := range options.ChannelLogLevelOverrides {
		m[k] = v
	}
	m[channel] = level
	options.ChannelLogLevelOverrides = m
}

func (options *Options) ClearChannelLogLevel(channel string) {
	m := map[string]logrus.Level{}
	for k, v := range options.ChannelLogLevelOverrides {
		if k != channel {
			m[k] = v
		}
	}
	options.ChannelLogLevelOverrides = m
}

func (options *Options) Color() *Options {
	options.PanicLabel = ansi.Red + "  PANIC" + ansi.DefaultFG
	options.FatalLabel = ansi.Red + "  FATAL" + ansi.DefaultFG
	options.ErrorLabel = ansi.Red + "  ERROR" + ansi.DefaultFG
	options.WarningLabel = ansi.Yellow + "WARNING" + ansi.DefaultFG
	options.InfoLabel = ansi.White + "   INFO" + ansi.DefaultFG
	options.DebugLabel = ansi.Blue + "  DEBUG" + ansi.DefaultFG
	options.TraceLabel = ansi.LightBlack + "  TRACE" + ansi.DefaultFG

	options.TimestampColor = ansi.Blue
	options.FunctionColor = ansi.Cyan
	options.FieldsColor = ansi.LightCyan
	options.DefaultFgColor = ansi.DefaultFG

	return options
}

func (options *Options) NoColor() *Options {
	options.PanicLabel = "  PANIC"
	options.FatalLabel = "  FATAL"
	options.ErrorLabel = "  ERROR"
	options.WarningLabel = "WARNING"
	options.InfoLabel = "   INFO"
	options.DebugLabel = "  DEBUG"
	options.TraceLabel = "  TRACE"

	options.TimestampColor = ""
	options.FunctionColor = ""
	options.FieldsColor = ""
	options.DefaultFgColor = ""

	return options
}

func CloneLogger(logger *logrus.Logger) *logrus.Logger {
	return &logrus.Logger{
		Out:          logger.Out,
		Hooks:        logger.Hooks,
		Formatter:    logger.Formatter,
		ReportCaller: logger.ReportCaller,
		Level:        logger.Level,
		ExitFunc:     logger.ExitFunc,
	}
}

func defaultEnv(env string, defaultValue bool) bool {
	if envStr := strings.ToLower(os.Getenv(env)); envStr != "" {
		if envValue, err := strconv.ParseBool(envStr); err == nil {
			return envValue
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "error parsing environment variable '%s' (%v)\n", env, err)
		}
	}
	return defaultValue
}
