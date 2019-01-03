//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"os"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/go-kit/kit/log"
)

const (
	PersistenceDB   = "database"
	PersistenceFile = "file"
)

type persistence interface {
	add(logEntry models.LogEntry) error
	closeSession()
	remove(criteria matchCriteria) (int, error)
	find(criteria matchCriteria) ([]models.LogEntry, error)

	// Needed for the tests. Reset the instance (closing files, sessions...)
	// and clear the logs.
	reset()
}

type privLogger struct {
	logLevel     *string
	rootLogger   log.Logger
	levelLoggers map[string]log.Logger
}

func newPrivateLogger() privLogger {
	pl := privLogger{}
	logLevel := logger.InfoLog
	pl.logLevel = &logLevel

	// Set up the loggers
	pl.levelLoggers = map[string]log.Logger{}

	pl.rootLogger = log.NewLogfmtLogger(os.Stdout)
	pl.rootLogger = log.WithPrefix(pl.rootLogger, "ts", log.DefaultTimestampUTC,
		"source", log.Caller(5))

	return pl
}

func (l privLogger) log(logLevel string, msg string, args ...interface{}) {
	// Check minimum log level
	for _, name := range logger.LogLevels {
		if name == *l.logLevel {
			break
		}
		if name == logLevel {
			return
		}
	}

	if dbClient != nil {
		logEntry := models.LogEntry{
			Level:         logLevel,
			Args:          args,
			OriginService: internal.SupportLoggingServiceKey,
			Message:       msg,
			Created:       db.MakeTimestamp(),
		}
		dbClient.add(logEntry)
	}

	if args == nil {
		args = []interface{}{"msg", msg}
	} else {
		if len(args)%2 == 1 {
			// add an empty string to keep k/v pairs correct
			args = append(args, "")
		}
		// Practical usage thus far has been to call this type like so Logger.Info("message")
		// I'm attempting to preserve that behavior below without requiring the client types
		// to provide the "msg" key.
		args = append(args, "msg", msg)
	}

	if l.levelLoggers[logLevel] == nil {
		l.levelLoggers[logLevel] = log.WithPrefix(l.rootLogger, "level", logLevel)
	}
	l.levelLoggers[logLevel].Log(args...)

}

// SetLogLevel sets logger log level
func (l privLogger) SetLogLevel(logLevel string) error {
	if logger.IsValidLogLevel(logLevel) == true {
		*l.logLevel = logLevel
		return nil
	}
	return types.ErrNotFound{}
}

func (l privLogger) Debug(msg string, args ...interface{}) {
	l.log(logger.DebugLog, msg, args...)
}

func (l privLogger) Error(msg string, args ...interface{}) {
	l.log(logger.ErrorLog, msg, args...)
}

func (l privLogger) Info(msg string, args ...interface{}) {
	l.log(logger.InfoLog, msg, args...)
}

func (l privLogger) Trace(msg string, args ...interface{}) {
	l.log(logger.TraceLog, msg, args...)
}

func (l privLogger) Warn(msg string, args ...interface{}) {
	l.log(logger.WarnLog, msg, args...)
}
