//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

// +build mongoRunning

// This test will only be executed if the tag mongoRunning is added when running
// the tests with a command like:
// go test -tags mongoRunning

package clients

import (
	"testing"
)

func TestMongoDB(t *testing.T) {

	t.Log("This test needs to have a running mongo on localhost")

	config := DBConfiguration{
		DbType:       MONGO,
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}

	mongo, err := newMongoClient(config)
	if err != nil {
		t.Fatalf("Could not connect with mongodb: %v", err)
	}

	testDB(t, mongo)
}
