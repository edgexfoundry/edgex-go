//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/models"
)

const (
	PersistenceMongo = "mongodb"
	PersistenceFile  = "file"
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
}

func (l privLogger) log(level string, msg string, labels []string) {
	if dbClient != nil {
		logEntry := models.LogEntry{
			Level:         level,
			Labels:        labels,
			OriginService: "logging",
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
