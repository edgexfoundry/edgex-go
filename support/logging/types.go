//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"github.com/edgexfoundry/edgex-go/support/domain"
)

const (
	PersistenceMongo = "mongodb"
	PersistenceFile  = "file"
)

type persistence interface {
	add(logEntry support_domain.LogEntry)
	remove(criteria matchCriteria) int
	find(criteria matchCriteria) []support_domain.LogEntry

	// Needed for the tests. Reset the instance (closing files, sessions...)
	// and clear the logs.
	reset()
}
