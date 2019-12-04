/*******************************************************************************
* Copyright 2019 Dell Inc.
*
* Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
* in compliance with the License. You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software distributed under the License
* is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
* or implied. See the License for the specific language governing permissions and limitations under
* the License.
*******************************************************************************/

// interfaces defines contracts that are externally available for use by other packages.
package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	PersistenceDB   = "database"
	PersistenceFile = "file"
)

// Persistence defines the minimum functionality required for any long lived logging repository.
type Persistence interface {
	// Add takes in a LogEntry, commits that data to the repository.
	// Returns an error if one exists, nil if successful.
	Add(logEntry models.LogEntry) error

	// CloseSession is an implementation agnostic way to shut down the repository and stop using its resources.
	CloseSession()

	// Remove takes in some Criteria about some LogEntry objects and removes them from the repository.
	// Returns the number of removed elements as an integer and an error if one exists, nil if successful.
	Remove(criteria Criteria) (int, error)

	// Find takes in some Criteria about some LogEntry objects and retrieves them from the repository.
	// Returns the matched elements as a slice of LogEntry and an error if one exists, nil if successful.
	Find(criteria Criteria) ([]models.LogEntry, error)
}
