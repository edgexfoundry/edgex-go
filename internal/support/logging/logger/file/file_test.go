//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package file

import (
	"os"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/support/logging/filter"

	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/stretchr/testify/assert"
)

func newSUT(filename string) *Logger {
	_ = os.Remove(filename)
	return NewLogger(filename)
}

var logEntries = []models.LogEntry{
	{Level: models.TraceLog, OriginService: "service", Message: "message1"},
	{Level: models.TraceLog, OriginService: "service", Message: "message2"},
	{Level: models.TraceLog, OriginService: "service", Message: "message1"},
	{Level: models.TraceLog, OriginService: "service", Message: "message2"},
	{Level: models.TraceLog, OriginService: "service", Message: "message1"},
}

func addLoggingData(t *testing.T, logger *Logger) int {
	for _, entry := range logEntries {
		assert.Nil(t, logger.Add(entry))
	}
	return len(logEntries)
}

func newSUTWithLogData(t *testing.T, filename string) (logger *Logger) {
	logger = newSUT(filename)
	addLoggingData(t, logger)
	return
}

func TestFileAdd(t *testing.T) {
	testFilename := "test_add.log"
	sut := newSUT(testFilename)

	logs, err := sut.Find(filter.Criteria{})
	assert.NotNil(t, err, "Error expected but not thrown")

	expectedEntries := addLoggingData(t, sut)

	logs, err = sut.Find(filter.Criteria{})
	assert.Nil(t, err, "Error thrown: %v", err)
	assert.NotNil(t, logs, "Should not be nil")
	assert.Equal(t, expectedEntries, len(logs), "Should return %d log entry, returned %d", expectedEntries, len(logs))

	sut.CloseSession()
	assert.Nil(t, os.Remove(testFilename))
}

func TestFileFind(t *testing.T) {
	var tests = []struct {
		name     string
		criteria filter.Criteria
		result   int
	}{
		{"empty", filter.Criteria{}, 5},
		{"keywords1", filter.Criteria{Keywords: []string{"1"}}, 3},
		{"keywords2", filter.Criteria{Keywords: []string{"2"}}, 2},
		{"keywords12", filter.Criteria{Keywords: []string{"2", "1"}}, 5},
	}

	testFilename := "test_find.log"
	sut := newSUTWithLogData(t, testFilename)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, err := sut.Find(tt.criteria)

			assert.Nil(t, err, "Error thrown: %v", err)
			assert.NotNil(t, logs, "Should not be nil")
			assert.Equal(t, tt.result, len(logs), "Should return %d log entries, returned %d", tt.result, len(logs))
		})
	}

	sut.CloseSession()
	assert.Nil(t, os.Remove(testFilename))
}

func TestFileRemove(t *testing.T) {
	var tests = []struct {
		name     string
		filename string
		criteria filter.Criteria
		result   int
	}{
		{"empty", "test_remove_1.log", filter.Criteria{}, 5},
		{"keywords1", "test_remove_2.log", filter.Criteria{Keywords: []string{"1"}}, 3},
		{"keywords2", "test_remove_3.log", filter.Criteria{Keywords: []string{"2"}}, 2},
		{"keywords12", "test_remove_4.log", filter.Criteria{Keywords: []string{"2", "1"}}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := newSUTWithLogData(t, tt.filename)

			removed, err := sut.Remove(tt.criteria)

			assert.Nil(t, err, "Error thrown: %v", err)
			assert.Equal(t, tt.result, removed, "Should return %d log entries, returned %d", tt.result, removed)

			sut.CloseSession()
			assert.Nil(t, os.Remove(tt.filename))
		})
	}
}
