//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

// +build mongoRunning

// This test will only be executed if the tag mongoRunning is added when running
// the tests with a command like:
// go test -tags mongoRunning

// Package mongo provides integration tests for interaction with MongoDO.
//
// Deprecated: Mongo functionality is deprecated as of the Geneva release and no new functionality should be added;
// Only bugs which address legacy issues
package mongo

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/test"
)

// TestMongoDB
//
// Deprecated: Mongo functionality is deprecated as of the Geneva release.
func TestMongoDB(t *testing.T) {

	t.Log("This test needs to have a running mongo on localhost")

	config := db.Configuration{
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}
	mongo, err := NewClient(config)
	if err != nil {
		t.Fatalf("Could not connect: %v", err)
	}
	test.TestDataDB(t, mongo)

	config.DatabaseName = "metadata"
	mongo, err = NewClient(config)
	if err != nil {
		t.Fatalf("Could not connect: %v", err)
	}
	test.TestMetadataDB(t, mongo)

	config.DatabaseName = "export"
	mongo, err = NewClient(config)
	if err != nil {
		t.Fatalf("Could not connect: %v", err)
	}
	test.TestExportDB(t, mongo)

	config.DatabaseName = "scheduler"
	mongo, err = NewClient(config)
	if err != nil {
		t.Fatalf("Could not connect: %v", err)
	}
	test.TestSchedulerDB(t, mongo)
}

// BenchmarkMongoDB
//
// Deprecated: Mongo functionality is deprecated as of the Geneva release.
func BenchmarkMongoDB(b *testing.B) {

	b.Log("This benchmark needs to have a running mongo on localhost")

	config := db.Configuration{
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}
	mongo, err := NewClient(config)

	if err != nil {
		b.Fatalf("Could not connect with mongodb: %v", err)
	}

	s := mongo.getSessionCopy()
	defer s.Close()

	test.BenchmarkDB(b, mongo)
}
