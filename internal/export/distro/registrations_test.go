//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/pkg/types"

	"github.com/google/uuid"
	"github.com/ugorji/go/codec"
)

func validRegistration() contract.Registration {
	r := contract.Registration{}
	r.Addressable = contract.Addressable{Id: uuid.New().String(), Name: "Test Addressable"}
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

func (sender *dummyStruct) Send(data []byte, ctx context.Context) bool {
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
	ri.processMessage(msgTypes.MessageEnvelope{})

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
	msg1 := msgTypes.MessageEnvelope{}
	msg1.ContentType = clients.ContentTypeJSON
	msg1.Payload, _ = json.Marshal(e1)
	ri.processMessage(msg1)

	e2 := &models.Event{}
	e2.Device = filterOutDev
	msg2 := msgTypes.MessageEnvelope{}
	msg2.ContentType = clients.ContentTypeJSON
	msg2.Payload, _ = json.Marshal(e2)
	ri.processMessage(msg2)

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

	// Assemble an event and MessageEnvelope for also testing CBOR handling
	e1 := models.Event{}
	e1.Device = "Some Device"
	e1.Readings = append(e1.Readings, contract.Reading{Name: "Reading1", Value: "ABC123"})

	var handle codec.CborHandle
	data := make([]byte, 0, 64)
	enc := codec.NewEncoderBytes(&data, &handle)
	err := enc.Encode(e1)
	if err != nil {
		t.Fatal("cbor error: " + err.Error())
	}
	msg1 := msgTypes.MessageEnvelope{ContentType: clients.ContentTypeCBOR, CorrelationID: uuid.New().String(),
		Payload: data, Checksum: "1234567890"}

	go func() {
		ri.chMessages <- msgTypes.MessageEnvelope{}
		ri.chMessages <- msg1
		ri.chRegistration <- nil
	}()
	ri.format = &dummyStruct{}
	ri.sender = &dummyStruct{}
	ri.encrypt = &dummyStruct{}
	ri.compression = &dummyStruct{}
	ri.filter = nil
	// Process two events (one JSON, one CBOR) and terminate
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

	msg := msgTypes.MessageEnvelope{}
	msg.ContentType = clients.ContentTypeJSON
	msg.Payload, _ = json.Marshal(event)

	ri := newRegistrationInfo()
	Dummy.count = 0

	ri.format = Dummy
	ri.sender = Dummy
	ri.encrypt = Dummy
	ri.compression = Dummy
	ri.filter = nil

	b.Run("nil", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ri.processMessage(msg)
		}
		b.SetBytes(int64(Dummy.lastSize))
	})

	ri.format = jsonFormatter{}
	ri.compression = &gzipTransformer{}

	b.Run("json_gzip", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ri.processMessage(msg)
		}
		b.SetBytes(int64(Dummy.lastSize))
	})
}
