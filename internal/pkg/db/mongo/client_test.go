//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

// +build mongoRunning

// This test will only be executed if the tag mongoRunning is added when running
// the tests with a command like:
// go test -tags mongoRunning

package mongo

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/test"
)

func TestMongoDB(t *testing.T) {

	t.Log("This test needs to have a running mongo on localhost")

	config := db.Configuration{
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}

	mongo := NewClient(config)
	test.TestDataDB(t, mongo)

	config.DatabaseName = "metadata"
	mongo = NewClient(config)

	err := mongo.Connect()
	if err != nil {
		t.Fatalf("Could not connect with mongodb: %v", err)
	}

	s := mongo.getSessionCopy()
	defer s.Close()

	_, err = s.DB(mongo.database.Name).C(db.Addressable).RemoveAll(nil)
	if err != nil {
		t.Fatalf("Error removing previous data: %v", err)
	}

	_, err = s.DB(mongo.database.Name).C(db.DeviceService).RemoveAll(nil)
	if err != nil {
		t.Fatalf("Error removing previous data: %v", err)
	}
	_, err = s.DB(mongo.database.Name).C(db.DeviceProfile).RemoveAll(nil)
	if err != nil {
		t.Fatalf("Error removing previous data: %v", err)
	}
	_, err = s.DB(mongo.database.Name).C(db.DeviceReport).RemoveAll(nil)
	if err != nil {
		t.Fatalf("Error removing previous data: %v", err)
	}
	_, err = s.DB(mongo.database.Name).C(db.ScheduleEvent).RemoveAll(nil)
	if err != nil {
		t.Fatalf("Error removing previous data: %v", err)
	}
	_, err = s.DB(mongo.database.Name).C(db.Device).RemoveAll(nil)
	if err != nil {
		t.Fatalf("Error removing previous data: %v", err)
	}
	_, err = s.DB(mongo.database.Name).C(db.ProvisionWatcher).RemoveAll(nil)
	if err != nil {
		t.Fatalf("Error removing previous data: %v", err)
	}

	test.TestMetadataDB(t, mongo)
}

func BenchmarkMongoDB(b *testing.B) {

	b.Log("This benchmark needs to have a running mongo on localhost")

	config := db.Configuration{
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}
	mongo := NewClient(config)

	test.BenchmarkDB(b, mongo)
}
