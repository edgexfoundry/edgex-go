//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
)

func validRegistration() contract.Registration {
	r := contract.Registration{}
	r.Format = contract.FormatJSON
	r.Compression = contract.CompNone
	r.Destination = contract.DestMQTT
	r.Encryption.Algo = contract.EncNone
	r.Filter.DeviceIDs = append(r.Filter.DeviceIDs, "dummy1")
	r.Filter.ValueDescriptorIDs = append(r.Filter.DeviceIDs, "dummy1")
	return r
}

func TestRegistrationInfoUpdate(t *testing.T) {
	ri := newRegistrationInfo()
	if ri == nil {
		t.Fatal("RegistrationInfo should not be nil")
	}

	r := contract.Registration{}
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
	count    int
	lastSize int
}

func (sender *dummyStruct) Send(data []byte, event *models.Event) bool {
	sender.count += 1
	sender.lastSize = len(data)

	return true
}

func (sender *dummyStruct) Format(ev *contract.Event) []byte {
	return []byte("")
}

func (sender *dummyStruct) Transform(data []byte) []byte {
	return data
}

func TestRegistrationInfoEvent(t *testing.T) {
	const (
		dummyDev     = "dummyDev"
		filterOutDev = "filterOutDev"
	)

	ri := newRegistrationInfo()
	// no configured should not panic
	ri.processEvent(&models.Event{})

	dummy := &dummyStruct{}

	ri.format = dummy
	ri.sender = dummy
	ri.encrypt = dummy
	ri.compression = dummy

	// Filter only accepting events from dummyDev
	f := contract.Filter{}
	f.DeviceIDs = append(f.DeviceIDs, dummyDev)
	filter := newDevIdFilter(f)

	ri.filter = append(ri.filter, filter)

	e1 := &models.Event{}
	e1.Device = dummyDev
	ri.processEvent(e1)

	e2 := &models.Event{}
	e2.Device = filterOutDev
	ri.processEvent(e2)

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
	if !ri.deleteFlag {
		t.Fatal("deleteme flag should be enabled after an invalid registration")
	}

	go func() {
		ri.chEvent <- &models.Event{}
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

func TestUpdateRunningRegistrations(t *testing.T) {
	running := make(map[string]*registrationInfo)

	if updateRunningRegistrations(running, contract.NotifyUpdate{}) == nil {
		t.Error("Err should not be nil")
	}
	if updateRunningRegistrations(running, contract.NotifyUpdate{
		Operation: contract.NotifyUpdateDelete}) == nil {
		t.Error("Err should not be nil")
	}
	if updateRunningRegistrations(running, contract.NotifyUpdate{
		Operation: contract.NotifyUpdateUpdate}) == nil {
		t.Error("Err should not be nil")
	}
	if updateRunningRegistrations(running, contract.NotifyUpdate{
		Operation: contract.NotifyUpdateAdd}) == nil {
		t.Error("Err should not be nil")
	}

}

func BenchmarkProcessEvent(b *testing.B) {
	var Dummy = &dummyStruct{}

	event := models.Event{}
	event.Device = "dummyDev"

	ri := newRegistrationInfo()
	Dummy.count = 0

	ri.format = Dummy
	ri.sender = Dummy
	ri.encrypt = Dummy
	ri.compression = Dummy
	ri.filter = nil

	b.Run("nil", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ri.processEvent(&event)
		}
		b.SetBytes(int64(Dummy.lastSize))
	})

	ri.format = jsonFormatter{}
	ri.compression = &gzipTransformer{}

	b.Run("json_gzip", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ri.processEvent(&event)
		}
		b.SetBytes(int64(Dummy.lastSize))
	})
}
