//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"github.com/edgexfoundry/edgex-go/internal/support/logging/interfaces"
	"os"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	testFilename   string = "test.log"
	sampleService1 string = "tservice1"
	sampleService2 string = "tservice2"
	message1       string = "message1"
	message2       string = "message2"
)

func testPersistenceFind(t *testing.T, persistence interfaces.Persistence) {
	var keywords1 = []string{"1"}
	var keywords2 = []string{"2"}
	var keywords12 = []string{"2", "1"}

	var tests = []struct {
		name     string
		criteria MatchCriteria
		result   int
	}{
		{"empty", MatchCriteria{}, 5},
		{"keywords1", MatchCriteria{Keywords: keywords1}, 3},
		{"keywords2", MatchCriteria{Keywords: keywords2}, 2},
		{"keywords12", MatchCriteria{Keywords: keywords12}, 5},
	}

	le := models.LogEntry{
		Level:         models.TraceLog,
		OriginService: sampleService1,
		Message:       message1,
	}
	persistence.add(le)
	le.Message = message2
	persistence.add(le)
	le.Message = message1
	persistence.add(le)
	le.Message = message2
	le.OriginService = sampleService2
	persistence.add(le)
	le.Message = message1
	persistence.add(le)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, err := persistence.find(tt.criteria)
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

	fl := FileLog{filename: testFilename}
	testPersistenceFind(t, &fl)
}

func testPersistenceRemove(t *testing.T, persistence interfaces.Persistence) {
	var keywords1 = []string{"1"}
	var keywords2 = []string{"2"}
	var keywords12 = []string{"2", "1"}

	var tests = []struct {
		name     string
		criteria MatchCriteria
		result   int
	}{
		{"empty", MatchCriteria{}, 5},
		{"keywords1", MatchCriteria{Keywords: keywords1}, 3},
		{"keywords2", MatchCriteria{Keywords: keywords2}, 2},
		{"keywords12", MatchCriteria{Keywords: keywords12}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			persistence.reset()

			le := models.LogEntry{
				Level:         models.TraceLog,
				OriginService: sampleService1,
				Message:       message1,
			}
			persistence.add(le)
			le.Message = message2
			persistence.add(le)
			le.Message = message1
			persistence.add(le)
			le.Message = message2
			le.OriginService = sampleService2
			persistence.add(le)
			le.Message = message1
			persistence.add(le)

			removed, err := persistence.remove(tt.criteria)
			if err != nil {
				t.Errorf("Error thrown: %s", err.Error())
			}
			if removed != tt.result {
				t.Errorf("Should return %d log entries, returned %d",
					tt.result, removed)
			}
			// we Add a new log
			persistence.add(le)
			logs, err := persistence.find(MatchCriteria{})
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

	fl := FileLog{filename: testFilename}
	testPersistenceRemove(t, &fl)
}
