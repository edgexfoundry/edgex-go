//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package file

import (
	"os"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/support/logging/criteria"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	testFilename   string = "test.log"
	sampleService1 string = "tservice1"
	sampleService2 string = "tservice2"
	message1       string = "message1"
	message2       string = "message2"
)

func testPersistenceFind(t *testing.T, persistence interfaces.Logger) {
	var keywords1 = []string{"1"}
	var keywords2 = []string{"2"}
	var keywords12 = []string{"2", "1"}

	var tests = []struct {
		name     string
		criteria criteria.Criteria
		result   int
	}{
		{"empty", criteria.Criteria{}, 5},
		{"keywords1", criteria.Criteria{Keywords: keywords1}, 3},
		{"keywords2", criteria.Criteria{Keywords: keywords2}, 2},
		{"keywords12", criteria.Criteria{Keywords: keywords12}, 5},
	}

	le := models.LogEntry{
		Level:         models.TraceLog,
		OriginService: sampleService1,
		Message:       message1,
	}
	persistence.Add(le)
	le.Message = message2
	persistence.Add(le)
	le.Message = message1
	persistence.Add(le)
	le.Message = message2
	le.OriginService = sampleService2
	persistence.Add(le)
	le.Message = message1
	persistence.Add(le)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, err := persistence.Find(tt.criteria)
			if err != nil {
				t.Errorf("Error thrown: %s", err.Error())
			}
			if logs == nil {
				t.Errorf("Should not be nil")
			}
			if len(logs) != tt.result {
				t.Errorf("Should return %d log entries, returned %d",
					tt.result, len(logs))
			}
		})
	}
}

func TestFileFind(t *testing.T) {
	// Remove test log, the test needs an empty file
	os.Remove(testFilename)

	// Remove test log when test ends
	defer os.Remove(testFilename)

	testPersistenceFind(t, NewLogger(testFilename))
}

func testPersistenceRemove(t *testing.T, persistence interfaces.Logger) {
	var keywords1 = []string{"1"}
	var keywords2 = []string{"2"}
	var keywords12 = []string{"2", "1"}

	var tests = []struct {
		name     string
		criteria criteria.Criteria
		result   int
	}{
		{"empty", criteria.Criteria{}, 5},
		{"keywords1", criteria.Criteria{Keywords: keywords1}, 3},
		{"keywords2", criteria.Criteria{Keywords: keywords2}, 2},
		{"keywords12", criteria.Criteria{Keywords: keywords12}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			persistence.Reset()

			le := models.LogEntry{
				Level:         models.TraceLog,
				OriginService: sampleService1,
				Message:       message1,
			}
			persistence.Add(le)
			le.Message = message2
			persistence.Add(le)
			le.Message = message1
			persistence.Add(le)
			le.Message = message2
			le.OriginService = sampleService2
			persistence.Add(le)
			le.Message = message1
			persistence.Add(le)

			removed, err := persistence.Remove(tt.criteria)
			if err != nil {
				t.Errorf("Error thrown: %s", err.Error())
			}
			if removed != tt.result {
				t.Errorf("Should return %d log entries, returned %d",
					tt.result, removed)
			}
			// we Add a new log
			persistence.Add(le)
			logs, err := persistence.Find(criteria.Criteria{})
			if len(logs) != 5-tt.result+1 {
				t.Errorf("Should return %d log entries, returned %d",
					6-tt.result+1, len(logs))
			}
		})
	}
}

func TestFileRemove(t *testing.T) {
	// Remove test log, the test needs an empty file
	os.Remove(testFilename)

	// Remove test log when test ends
	defer os.Remove(testFilename)

	testPersistenceRemove(t, NewLogger(testFilename))
}
