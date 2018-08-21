//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package models

import (
	"testing"

	"go.uber.org/zap"
)

const (
	deviceID1 = "DEV1"
	deviceID2 = "DEV2"

	descriptor1 = "Descriptor1"
	descriptor2 = "Descriptor2"
)

func TestFilterDevice(t *testing.T) {
	logger = zap.NewNop()
	defer logger.Sync()

	// Filter only accepting events from device 1
	f := Filter{}
	f.DeviceIDs = append(f.DeviceIDs, "DEV1")

	// Event from device 1
	eventDev1 := Event{
		Device: deviceID1,
	}
	// Event from device 2
	eventDev2 := Event{
		Device: deviceID2,
	}

	filter := NewDevIdFilter(f)
	accepted, _ := filter.Filter(nil)
	if accepted {
		t.Fatal("Event should be filtered out")
	}
	accepted, res := filter.Filter(&eventDev1)
	if !accepted {
		t.Fatal("Event should be accepted")
	}
	if res != &eventDev1 {
		t.Fatal("Event should be the same")
	}
	accepted, _ = filter.Filter(&eventDev2)
	if accepted {
		t.Fatal("Event should be filtered")
	}
}

func TestFilterValue(t *testing.T) {
	logger = zap.NewNop()
	defer logger.Sync()

	f1 := Filter{}
	f1.ValueDescriptorIDs = append(f1.ValueDescriptorIDs, descriptor1)

	f12 := Filter{}
	f12.ValueDescriptorIDs = append(f12.ValueDescriptorIDs, descriptor1)
	f12.ValueDescriptorIDs = append(f12.ValueDescriptorIDs, descriptor2)

	// only accepts value descriptor 1
	filter1 := NewValueDescFilter(f1)
	// accepts value descriptor 1 and 2
	filter12 := NewValueDescFilter(f12)

	// event with a value descriptor 1
	event1 := Event{}
	event1.Readings = append(event1.Readings, Reading{Name: descriptor1})

	// event with a value descriptor 2
	event2 := Event{}
	event2.Readings = append(event2.Readings, Reading{Name: descriptor2})

	// event with a value descriptor 1 and another 2
	event12 := Event{}
	event12.Readings = append(event12.Readings, Reading{Name: descriptor1})
	event12.Readings = append(event12.Readings, Reading{Name: descriptor2})

	accepted, res := filter1.Filter(nil)
	if accepted {
		t.Fatal("Event should be filtered out")
	}

	accepted, res = filter1.Filter(&event1)
	if !accepted {
		t.Fatal("Event should be accepted")
	}
	if len(res.Readings) != 1 {
		t.Fatal("Event should be one reading, there are ", len(res.Readings))
	}

	accepted, res = filter1.Filter(&event12)
	if !accepted {
		t.Fatal("Event should be accepted")
	}
	if len(res.Readings) != 1 {
		t.Fatal("Event should be one reading, there are ", len(res.Readings))
	}

	accepted, res = filter1.Filter(&event2)
	if accepted {
		t.Fatal("Event should be filtered out")
	}

	accepted, res = filter12.Filter(&event1)
	if !accepted {
		t.Fatal("Event should be accepted")
	}
	if len(res.Readings) != 1 {
		t.Fatal("Event should be one reading, there are ", len(res.Readings))
	}

	accepted, res = filter12.Filter(&event12)
	if !accepted {
		t.Fatal("Event should be accepted")
	}
	if len(res.Readings) != 2 {
		t.Fatal("Event should be one reading, there are ", len(res.Readings))
	}

	accepted, res = filter12.Filter(&event2)
	if !accepted {
		t.Fatal("Event should be accepted")
	}
	if len(res.Readings) != 1 {
		t.Fatal("Event should be one reading, there are ", len(res.Readings))
	}
}
