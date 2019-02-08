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
	"strconv"
	"strings"
	"testing"

	contract "github.com/edgexfoundry/edgex-go/pkg/models"
)

const (
	devID1        = "id1"
	readingName1  = "sensor1"
	readingValue1 = "123.45"
)

func TestJson(t *testing.T) {
	eventIn := contract.Event{
		Device: devID1,
	}

	jf := jsonFormatter{}
	out := jf.Format(&eventIn)
	if out == nil {
		t.Fatal("out should not be nil")
	}

	var eventOut contract.Event
	if err := json.Unmarshal(out, &eventOut); err != nil {
		t.Fatalf("Error unmarshalling event: %v", err)
	}
	if !reflect.DeepEqual(eventIn, eventOut) {
		t.Fatalf("Objects should be equals: %v %v", eventIn, eventOut)
	}
}

func TestXml(t *testing.T) {
	eventIn := contract.Event{
		Device: devID1,
	}

	xf := xmlFormatter{}
	out := xf.Format(&eventIn)
	if out == nil {
		t.Fatal("out should not be nil")
	}

	var eventOut contract.Event
	if err := xml.Unmarshal(out, &eventOut); err != nil {
		t.Fatalf("Error unmarshalling event: %v", err)
	}
	if !reflect.DeepEqual(eventIn, eventOut) {
		t.Fatalf("Objects should be equals: %v %v", eventIn, eventOut)
	}
}

func TestThingsBoardJson(t *testing.T) {
	eventIn := contract.Event{
		Device: devID1,
	}

	tbjf := thingsboardJSONFormatter{}
	out := tbjf.Format(&eventIn)
	if out == nil {
		t.Fatal("out should not be nil")
	}

	s := string(out[:])
	if strings.HasPrefix(s, "{\""+devID1+"\":[{\"ts\":") == false {
		t.Fatalf("Invalid ThingsBoard JSON format: %v", s)
	}
}

func TestNoop(t *testing.T) {
	eventIn := contract.Event{
		Device: devID1,
	}

	xf := noopFormatter{}
	out := xf.Format(&eventIn)

	if out == nil {
		t.Fatal("out should not be nil")
	}

	if len(out) != 0 {
		t.Fatal("Formmated array length is not zero, length = " + strconv.Itoa(len(out)))
	}
}

func TestAWSIoTJson(t *testing.T) {
	eventIn := contract.Event{}

	eventIn.Readings = append(eventIn.Readings, contract.Reading{Device: devID1, Name: readingName1, Value: readingValue1})

	af := awsFormatter{}
	out := af.Format(&eventIn)

	if out == nil {
		t.Fatal("out should not be nil")
	}

	var sd interface{}
	err := json.Unmarshal(out, &sd)

	if err != nil {
		t.Fatalf("Error unmarshal the formatted string: %v %v", err, out)
	}

	shadow := sd.(map[string]interface{})

	state := shadow["state"].(map[string]interface{})

	reported := state["reported"].(map[string]interface{})

	val, err := strconv.ParseFloat(readingValue1, 64)

	if reported[readingName1] != val {
		t.Fatalf("Unmshalred json is not correct: %v", reported)
	}
}

func TestBIoT(t *testing.T) {
	eventIn := contract.Event{}

	eventIn.Readings = append(eventIn.Readings, contract.Reading{Device: devID1, Name: readingName1, Value: readingValue1})

	xf := biotFormatter{}
	out := xf.Format(&eventIn)

	if out == nil {
		t.Fatal("out should not be nil")
	}

	var sd interface{}
	err := json.Unmarshal(out, &sd)

	if err != nil {
		t.Fatalf("Error unmarshal the formatted string: %v %v", err, out)
	}
}
