//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import "github.com/edgexfoundry/edgex-go/internal/support/logging/models"

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
