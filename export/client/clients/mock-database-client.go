/*******************************************************************************
 * Copyright 2018 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * @author: Trevor Conn, Dell
 * @version: 0.5.0
 *******************************************************************************/
package clients

import (
	"time"

	"github.com/edgexfoundry/edgex-go/export"
	"gopkg.in/mgo.v2/bson"
)

type MockParams struct {
	RegistrationId bson.ObjectId
	Name           string
	Origin         int64
}

var mockParams *MockParams

func NewMockParams() *MockParams {
	if mockParams == nil {
		mockParams = &MockParams{
			RegistrationId: bson.NewObjectId(),
			Name:           "test export",
			Origin:         123456789}
	}
	return mockParams
}

type MockDb struct {
}

func (mc *MockDb) Registrations() ([]export.Registration, error) {
	ticks := time.Now().Unix()
	regs := []export.Registration{}

	reg1 := export.Registration{ID: mockParams.RegistrationId, Name: mockParams.Name, Created: ticks, Modified: ticks,
		Origin: mockParams.Origin, Format: export.FormatJSON, Compression: export.CompNone, Enable: true, Destination: export.DestMQTT}

	regs = append(regs, reg1)
	return regs, nil
}

func (mc *MockDb) AddRegistration(reg *export.Registration) (bson.ObjectId, error) {
	reg.ID = bson.NewObjectId()
	return reg.ID, nil
}

func (mc *MockDb) UpdateRegistration(reg export.Registration) error {
	return nil
}

func (mc *MockDb) RegistrationById(id string) (export.Registration, error) {
	ticks := time.Now().Unix()

	if id == mockParams.RegistrationId.Hex() {
		return export.Registration{ID: mockParams.RegistrationId, Name: mockParams.Name, Created: ticks, Modified: ticks,
			Origin: mockParams.Origin, Format: export.FormatJSON, Compression: export.CompNone, Enable: true, Destination: export.DestMQTT}, nil
	}
	return export.Registration{}, ErrNotFound
}

func (mc *MockDb) RegistrationByName(name string) (export.Registration, error) {
	ticks := time.Now().Unix()

	if name == mockParams.Name {
		return export.Registration{ID: mockParams.RegistrationId, Name: mockParams.Name, Created: ticks, Modified: ticks,
			Origin: mockParams.Origin, Format: export.FormatJSON, Compression: export.CompNone, Enable: true, Destination: export.DestMQTT}, nil
	}
	return export.Registration{}, ErrNotFound
}

func (mc *MockDb) DeleteRegistrationById(id string) error {
	return nil
}

func (mc *MockDb) DeleteRegistrationByName(name string) error {
	return nil
}
