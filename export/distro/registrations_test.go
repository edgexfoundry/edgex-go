//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"github.com/edgexfoundry/edgex-go/export"

	// "go.uber.org/zap"
	"testing"
)

func validRegistration() export.Registration {
	r := export.Registration{}
	r.Format = export.FormatJSON
	r.Compression = export.CompNone
	r.Destination = export.DestMQTT
	r.Encryption.Algo = export.EncNone
	r.Filter.DeviceIDs = append(r.Filter.DeviceIDs, "dummy1")
	r.Filter.ValueDescriptorIDs = append(r.Filter.DeviceIDs, "dummy1")
	return r
}

func TestRegistrationInfoUpdate(t *testing.T) {
	ri := newRegistrationInfo()
	if ri == nil {
		t.Fatal("RegistrationInfo should not be nil")
	}

	r := export.Registration{}
	if ri.update(r) {
		t.Fatal("An empty registration is not valid")
	}

	r = validRegistration()
	if !ri.update(r) {
		t.Fatal("This registration should be good")
	}

	r.Format = "invalid"
	if ri.update(r) {
		t.Fatal("Registration with invalid fields")
	}

	r = validRegistration()
	r.Compression = "invalid"
	if ri.update(r) {
		t.Fatal("Registration with invalid fields")
	}

	r = validRegistration()
	r.Destination = "invalid"
	if ri.update(r) {
		t.Fatal("Registration with invalid fields")
	}

	r = validRegistration()
	r.Encryption.Algo = "invalid"
	if ri.update(r) {
		t.Fatal("Registration with invalid fields")
	}
}

type dummyStruct struct {
	count int
}

func (sender *dummyStruct) Send(data []byte) {
	sender.count += 1
}

func (sender *dummyStruct) Format(ev *export.Event) []byte {
	return []byte("")
}

func (sender *dummyStruct) Transform(data []byte) []byte {
	return data
}

func TestRegistrationInfoEvent(t *testing.T) {
	ri := newRegistrationInfo()
	// no configured should not panic
	ri.processEvent(&export.Event{})

	dummy := &dummyStruct{}

	ri.format = dummy
	ri.sender = dummy
	ri.encrypt = dummy
	ri.compression = dummy
	ri.filter = nil
	ri.processEvent(&export.Event{})
	if dummy.count != 1 {
		t.Fatal("It should send an event")
	}
}

func TestRegistrationInfoLoop(t *testing.T) {
	ri := newRegistrationInfo()
	ri.update(validRegistration())

	ri.format = &dummyStruct{}
	ri.sender = &dummyStruct{}
	ri.encrypt = &dummyStruct{}
	ri.compression = &dummyStruct{}
	ri.filter = nil

	go func() {
		ri.chRegistration <- nil
	}()
	// End loop receiving a nil update registration
	registrationLoop(ri)

	go func() {
		r := validRegistration()
		ri.chRegistration <- &r
		ri.chRegistration <- nil
	}()
	// update registration
	ri.filter = nil
	registrationLoop(ri)
	if len(ri.filter) != 2 {
		t.Fatal("There should be two filters after updating registration")
	}

	go func() {
		r := validRegistration()
		r.Compression = "INVALID"
		ri.chRegistration <- &r
	}()
	// update invalid registration,
	ri.filter = nil
	registrationLoop(ri)
	if !ri.deleteMe {
		t.Fatal("deleteme flag should be enabled after an invalid registration")
	}

	go func() {
		ri.chEvent <- &export.Event{}
		ri.chRegistration <- nil
	}()
	ri.format = &dummyStruct{}
	ri.sender = &dummyStruct{}
	ri.encrypt = &dummyStruct{}
	ri.compression = &dummyStruct{}
	ri.filter = nil
	// Process an event and terminate
	registrationLoop(ri)
}
