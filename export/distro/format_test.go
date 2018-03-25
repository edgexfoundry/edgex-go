//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"encoding/json"
	"encoding/xml"
	"reflect"
	"testing"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
)

const (
	devID1 = "id1"
)

func TestJson(t *testing.T) {
	eventIn := models.Event{
		Device: devID1,
	}

	jf := jsonFormater{}
	out := jf.Format(&eventIn)
	if out == nil {
		t.Fatal("out should not be nil")
	}

	var eventOut models.Event
	if err := json.Unmarshal(out, &eventOut); err != nil {
		t.Fatalf("Error unmarshalling event: %v", err)
	}
	if !reflect.DeepEqual(eventIn, eventOut) {
		t.Fatalf("Objects should be equals: %v %v", eventIn, eventOut)
	}
}

func TestXml(t *testing.T) {
	eventIn := models.Event{
		Device: devID1,
	}

	xf := xmlFormater{}
	out := xf.Format(&eventIn)
	if out == nil {
		t.Fatal("out should not be nil")
	}

	var eventOut models.Event
	if err := xml.Unmarshal(out, &eventOut); err != nil {
		t.Fatalf("Error unmarshalling event: %v", err)
	}
	if !reflect.DeepEqual(eventIn, eventOut) {
		t.Fatalf("Objects should be equals: %v %v", eventIn, eventOut)
	}
}
