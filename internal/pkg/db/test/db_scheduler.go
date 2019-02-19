//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package test

import (
	"fmt"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
)

func TestSchedulerDB(t *testing.T, db interfaces.DBClient) {
	testDBInterval(t, db)
	testDBIntervalAction(t, db)

	db.CloseSession()
	// Calling CloseSession twice to test that there is no panic when closing an
	// already closed db
	db.CloseSession()
}

func populateIntervals(db interfaces.DBClient, count int) (string, error) {
	var id string
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("name%d", i)
		mi := contract.Interval{}
		mi.Name = name
		var err error
		id, err = db.AddInterval(mi)
		if err != nil {
			return id, err
		}
	}
	return id, nil
}

func populateIntervalActions(db interfaces.DBClient, count int) (string, error) {
	var id string
	for i := 0; i < count; i++ {
		mi := contract.IntervalAction{}
		mi.Name = fmt.Sprintf("name%d", i)
		mi.Target = fmt.Sprintf("target%d", i)
		mi.Interval = fmt.Sprintf("interval%d", i)
		var err error
		id, err = db.AddIntervalAction(mi)
		if err != nil {
			return id, err
		}
	}
	return id, nil
}

func testDBInterval(t *testing.T, db interfaces.DBClient) {
	_, err := db.ScrubAllIntervals()
	if err != nil {
		t.Fatalf("Error removing all intervals")
	}

	mis, err := db.Intervals()
	if err != nil {
		t.Fatalf("Error getting intervals %v", err)
	}
	if len(mis) != 0 {
		t.Fatalf("There should be 0 mis instead of %d", len(mis))
	}

	id, err := populateIntervals(db, 110)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	_, err = populateIntervals(db, 110)
	if err == nil {
		t.Fatalf("Should be an error adding a new interval with the same name\n")
	}

	mis, err = db.Intervals()
	if err != nil {
		t.Fatalf("Error getting intervals %v", err)
	}
	if len(mis) != 110 {
		t.Fatalf("There should be 110 intervals instead of %d", len(mis))
	}

	mis, err = db.IntervalsWithLimit(50)
	if err != nil {
		t.Fatalf("Error getting intervals with limit: %v", err)
	}
	if len(mis) != 50 {
		t.Fatalf("There should be 50 intervals, not %d", len(mis))
	}

	mi, err := db.IntervalById(id)
	if err != nil {
		t.Fatalf("Error getting interval by id %v", err)
	}
	if mi.ID != id {
		t.Fatalf("Id does not match %s - %s", mi.ID, id)
	}

	_, err = db.IntervalById("INVALID")
	if err == nil {
		t.Fatalf("Interval should not be found")
	}

	mi, err = db.IntervalByName("name1")
	if err != nil {
		t.Fatalf("Error getting interval by id %v", err)
	}
	if mi.Name != "name1" {
		t.Fatalf("Name does not match %s - name1", mi.Name)
	}
	_, err = db.IntervalByName("INVALID")
	if err == nil {
		t.Fatalf("Interval should not be found")
	}

	mi = contract.Interval{}
	mi.ID = id
	mi.Name = "name"
	err = db.UpdateInterval(mi)
	if err != nil {
		t.Fatalf("Error updating interval %v", err)
	}
	mi2, err := db.IntervalById(mi.ID)
	if err != nil {
		t.Fatalf("Error getting interval by id %v", err)
	}
	if mi2.Name != mi.Name {
		t.Fatalf("Did not update interval correctly: %s %s", mi.Name, mi2.Name)
	}

	err = db.DeleteIntervalById("INVALID")
	if err == nil {
		t.Fatalf("Interval should not be deleted")
	}

	err = db.DeleteIntervalById(id)
	if err != nil {
		t.Fatalf("Interval should be deleted: %v", err)
	}

	err = db.UpdateInterval(mi)
	if err == nil {
		t.Fatalf("Update should return error")
	}

	_, err = db.ScrubAllIntervals()
	if err != nil {
		t.Fatalf("Error removing all intervals")
	}
}

func testDBIntervalAction(t *testing.T, db interfaces.DBClient) {
	_, err := db.ScrubAllIntervalActions()
	if err != nil {
		t.Fatalf("Error removing all IntervalActions")
	}

	ias, err := db.IntervalActions()
	if err != nil {
		t.Fatalf("Error getting IntervalActions %v", err)
	}
	if len(ias) != 0 {
		t.Fatalf("There should be 0 ias instead of %d", len(ias))
	}

	id, err := populateIntervalActions(db, 110)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	_, err = populateIntervalActions(db, 110)
	if err == nil {
		t.Fatalf("Should be an error adding a new IntervalAction with the same name\n")
	}

	ias, err = db.IntervalActions()
	if err != nil {
		t.Fatalf("Error getting IntervalActions %v", err)
	}
	if len(ias) != 110 {
		t.Fatalf("There should be 110 IntervalActions instead of %d", len(ias))
	}

	ias, err = db.IntervalActionsWithLimit(50)
	if err != nil {
		t.Fatalf("Error getting IntervalActions with limit: %v", err)
	}
	if len(ias) != 50 {
		t.Fatalf("There should be 50 IntervalActions, not %d", len(ias))
	}

	ia, err := db.IntervalActionById(id)
	if err != nil {
		t.Fatalf("Error getting IntervalAction by id %v", err)
	}
	if ia.ID != id {
		t.Fatalf("Id does not match %s - %s", ia.ID, id)
	}

	_, err = db.IntervalActionById("INVALID")
	if err == nil {
		t.Fatalf("IntervalAction should not be found")
	}

	ia, err = db.IntervalActionByName("name1")
	if err != nil {
		t.Fatalf("Error getting IntervalAction by id %v", err)
	}
	if ia.Name != "name1" {
		t.Fatalf("Name does not match %s - name1", ia.Name)
	}
	_, err = db.IntervalActionByName("INVALID")
	if err == nil {
		t.Fatalf("IntervalAction should not be found")
	}

	ias, err = db.IntervalActionsByTarget("target1")
	if err != nil {
		t.Fatalf("Error getting IntervalActionsByTarget: %v", err)
	}
	if len(ias) != 1 {
		t.Fatalf("There should be 1 IntervalActions, not %d", len(ias))
	}
	ias, err = db.IntervalActionsByTarget("INVALID")
	if err != nil {
		t.Fatalf("Error getting IntervalActionsByTarget: %v", err)
	}
	if len(ias) != 0 {
		t.Fatalf("There should be 0 IntervalActions, not %d", len(ias))
	}

	ias, err = db.IntervalActionsByIntervalName("interval1")
	if err != nil {
		t.Fatalf("Error getting IntervalActionsByIntervalName: %v", err)
	}
	if len(ias) != 1 {
		t.Fatalf("There should be 1 IntervalActions, not %d", len(ias))
	}
	ias, err = db.IntervalActionsByIntervalName("INVALID")
	if err != nil {
		t.Fatalf("Error getting IntervalActionsByIntervalName: %v", err)
	}
	if len(ias) != 0 {
		t.Fatalf("There should be 0 IntervalActions, not %d", len(ias))
	}

	ia = contract.IntervalAction{}
	ia.ID = id
	ia.Name = "name"
	err = db.UpdateIntervalAction(ia)
	if err != nil {
		t.Fatalf("Error updating IntervalAction %v", err)
	}
	ia2, err := db.IntervalActionById(ia.ID)
	if err != nil {
		t.Fatalf("Error getting IntervalAction by id %v", err)
	}
	if ia2.Name != ia.Name {
		t.Fatalf("Did not update IntervalAction correctly: %s %s", ia.Name, ia2.Name)
	}

	err = db.DeleteIntervalActionById("INVALID")
	if err == nil {
		t.Fatalf("IntervalAction should not be deleted")
	}

	err = db.DeleteIntervalActionById(id)
	if err != nil {
		t.Fatalf("IntervalAction should be deleted: %v", err)
	}

	err = db.UpdateIntervalAction(ia)
	if err == nil {
		t.Fatalf("Update should return error")
	}

	_, err = db.ScrubAllIntervalActions()
	if err != nil {
		t.Fatalf("Error removing all IntervalActions")
	}
}
