//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"fmt"
	"log"
	"os"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/models"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
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
	stdOutLogger *log.Logger
	logLevel     *string
}

func newPrivateLogger() privLogger {
	p := privLogger{}
	p.stdOutLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	logLevel := logger.InfoLog
	p.logLevel = &logLevel
	return p
}

func (l privLogger) log(logLevel string, msg string, labels []string) {
	// Check minimum log level
	for _, name := range logger.LogLevels {
		if name == *l.logLevel {
			break
		}
		if name == logLevel {
			return
		}
	}

	l.stdOutLogger.SetPrefix(fmt.Sprintf("%s: ", logLevel))
	l.stdOutLogger.Println(msg)
	if dbClient != nil {
		logEntry := models.LogEntry{
			Level:         logLevel,
			Labels:        labels,
			OriginService: internal.SupportLoggingServiceKey,
			Message:       msg,
			Created:       db.MakeTimestamp(),
		}
		dbClient.add(logEntry)
	}
}

// SetLogLevel sets logger log level
func (l privLogger) SetLogLevel(logLevel string) {
	if logger.IsValidLogLevel(logLevel) == true {
		*l.logLevel = logLevel
	}
}

func (l privLogger) Debug(msg string, labels ...string) error {
	l.log(logger.DebugLog, msg, labels)
	return nil
}

func (l privLogger) Error(msg string, labels ...string) error {
	l.log(logger.ErrorLog, msg, labels)
	return nil
}

func (l privLogger) Info(msg string, labels ...string) error {
	l.log(logger.InfoLog, msg, labels)
	return nil
}

func (l privLogger) Trace(msg string, labels ...string) error {
	l.log(logger.TraceLog, msg, labels)
	return nil
}

func (l privLogger) Warn(msg string, labels ...string) error {
	l.log(logger.WarnLog, msg, labels)
	return nil
}
