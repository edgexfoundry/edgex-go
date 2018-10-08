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
}

func newPrivateLogger() privLogger {
	p := privLogger{}
	p.stdOutLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	return p
}

func (l privLogger) log(level string, msg string, labels []string) {
	l.stdOutLogger.SetPrefix(fmt.Sprintf("%s: ", level))
	l.stdOutLogger.Println(msg)
	if dbClient != nil {
		logEntry := models.LogEntry{
			Level:         level,
			Labels:        labels,
			OriginService: internal.SupportLoggingServiceKey,
			Message:       msg,
			Created:       db.MakeTimestamp(),
		}
		dbClient.add(logEntry)
	}
}

func (l privLogger) Debug(msg string, labels ...string) error {
	l.log(models.DEBUG, msg, labels)
	return nil
}

func (l privLogger) Error(msg string, labels ...string) error {
	l.log(models.ERROR, msg, labels)
	return nil
}

func (l privLogger) Info(msg string, labels ...string) error {
	l.log(models.INFO, msg, labels)
	return nil
}

func (l privLogger) Trace(msg string, labels ...string) error {
	l.log(models.TRACE, msg, labels)
	return nil
}

func (l privLogger) Warn(msg string, labels ...string) error {
	l.log(models.WARN, msg, labels)
	return nil
}
