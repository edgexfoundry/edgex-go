//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
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

func TestFind(t *testing.T) {
	// Remove test log, the test needs an empty file
	_ = os.Remove(testFilename)

	// Remove test log when test ends
	defer os.Remove(testFilename)

	persistence := fileLog{filename: testFilename}

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
	_ = persistence.Add(le)
	le.Message = message2
	_ = persistence.Add(le)
	le.Message = message1
	_ = persistence.Add(le)
	le.Message = message2
	le.OriginService = sampleService2
	_ = persistence.Add(le)
	le.Message = message1
	_ = persistence.Add(le)

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

func TestRemove(t *testing.T) {
	// Remove test log, the test needs an empty file
	_ = os.Remove(testFilename)

	// Remove test log when test ends
	defer os.Remove(testFilename)

	persistence := fileLog{filename: testFilename}

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
			if persistence.out != nil {
				_ = persistence.out.Close()
				persistence.out = nil
			}
			_ = os.Remove(persistence.filename)

			le := models.LogEntry{
				Level:         models.TraceLog,
				OriginService: sampleService1,
				Message:       message1,
			}
			_ = persistence.Add(le)
			le.Message = message2
			_ = persistence.Add(le)
			le.Message = message1
			_ = persistence.Add(le)
			le.Message = message2
			le.OriginService = sampleService2
			_ = persistence.Add(le)
			le.Message = message1
			_ = persistence.Add(le)

			removed, err := persistence.Remove(tt.criteria)
			if err != nil {
				t.Errorf("Error thrown: %s", err.Error())
			}
			if removed != tt.result {
				t.Errorf("Should return %d log entries, returned %d",
					tt.result, removed)
			}
			// we add a new log
			_ = persistence.Add(le)
			logs, err := persistence.Find(MatchCriteria{})
			if len(logs) != 5-tt.result+1 {
				t.Errorf("Should return %d log entries, returned %d",
					6-tt.result+1, len(logs))
			}
		})
	}
}
