package logging

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Level       string    `json:"level"`
	Formatter   formatter `json:"formatter"`
	LocalLogger bool      `json:"localLogger"`
	AddSource   bool      `json:"addSource"`
}

type formatter struct {
	Format string                 `json:"format"`
	Data   map[string]interface{} `json:"data"`
}

type loggingConfig Config

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	log = (*logger)(logrus.New())
	err := unmarshal((*loggingConfig)(c))
	if err != nil {
		return err
	}
	return c.SetLogger()
}

func (c *Config) UnmarshalJSON(data []byte) error {
	log = (*logger)(logrus.New())
	err := json.Unmarshal(data, (*loggingConfig)(c))
	if err != nil {
		return err
	}
	return c.SetLogger()
}

func (c *Config) SetLogger() (err error) {
	err = c.parseFormatter()
	if err != nil {
		return err
	}
	err = c.parseLevel()
	if err != nil {
		return err
	}
	err = c.unmarshalFormatter()
	if err != nil {
		return err
	}
	c.setGlobal()
	return nil
}

func (c *Config) setGlobal() {
	if c.LocalLogger {
		return
	}
	logrus.SetFormatter(log.Formatter)
	logrus.SetLevel(log.Level)
	logrus.SetReportCaller(log.ReportCaller)
	log = (*logger)(logrus.StandardLogger())
}

func (c *Config) unmarshalFormatter() error {
	formatterData, err := json.Marshal(c.Formatter.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(formatterData, log.Formatter)
}

func (c *Config) parseLevel() error {
	if c.Level == "" {
		log.Level = logrus.InfoLevel
		return nil
	}
	level, err := logrus.ParseLevel(c.Level)
	if err != nil {
		return err
	}
	log.Level = level
	return nil
}

const (
	FormatterText = "text"
	FormatterJSON = "json"
)

func (c *Config) parseFormatter() error {
	switch c.Formatter.Format {
	case FormatterJSON:
		log.Formatter = &logrus.JSONFormatter{}
	case FormatterText, "":
		log.Formatter = &logrus.TextFormatter{}
	default:
		return fmt.Errorf("%s formatter not supported", c.Formatter)
	}
	return nil
}

// Slog constructs a slog.Logger with the Formatter and Level from config.
func (c *Config) Slog() *slog.Logger {
	logger := slog.Default()

	var level slog.Level
	if err := level.UnmarshalText([]byte(c.Level)); err != nil {
		logger.Warn("invalid config, using default slog", "err", err)
		return logger
	}
	opts := &slog.HandlerOptions{
		AddSource:   c.AddSource,
		Level:       level,
		ReplaceAttr: c.fieldMapToPlaceKey(),
	}

	switch c.Formatter.Format {
	case FormatterText:
		return slog.New(slog.NewTextHandler(os.Stderr, opts))
	case FormatterJSON:
		return slog.New(slog.NewJSONHandler(os.Stderr, opts))
	case "":
		logger.Warn("no slog format in config, using text handler")
	default:
		logger.Warn("unknown slog format in config, using text handler", "format", c.Formatter.Format)
	}
	return slog.New(slog.NewTextHandler(os.Stderr, opts))
}

func (c *Config) fieldMapToPlaceKey() func(groups []string, a slog.Attr) slog.Attr {
	fieldMap, ok := c.Formatter.Data["fieldmap"].(map[string]interface{})
	if !ok {
		return nil
	}
	return func(groups []string, a slog.Attr) slog.Attr {
		for key, newKey := range fieldMap {
			if a.Key == key {
				a.Key = newKey.(string)
				return a
			}
		}
		return a
	}
}
