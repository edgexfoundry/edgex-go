/*******************************************************************************
 * Copyright 2018
 * Dell Inc.
 * Cavium
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
 *******************************************************************************/
package memory

import (
	"time"

	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/globalsign/mgo/bson"
)

func (mc *MemDB) Registrations() ([]export.Registration, error) {
	return mc.regs, nil
}

func (mc *MemDB) AddRegistration(reg *export.Registration) (bson.ObjectId, error) {
	ticks := time.Now().Unix()
	reg.Created = ticks
	reg.Modified = ticks
	reg.ID = bson.NewObjectId()

	mc.regs = append(mc.regs, *reg)

	return reg.ID, nil
}

func (mc *MemDB) UpdateRegistration(reg export.Registration) error {
	for i, r := range mc.regs {
		if r.ID == reg.ID {
			mc.regs[i] = reg
			return nil
		}
	}
	return db.ErrNotFound
}

func (mc *MemDB) RegistrationById(id string) (export.Registration, error) {
	for _, reg := range mc.regs {
		if reg.ID.Hex() == id {
			return reg, nil
		}
	}

	return export.Registration{}, db.ErrNotFound
}

func (mc *MemDB) RegistrationByName(name string) (export.Registration, error) {
	for _, reg := range mc.regs {
		if reg.Name == name {
			return reg, nil
		}
	}

	return export.Registration{}, db.ErrNotFound
}

func (mc *MemDB) DeleteRegistrationById(id string) error {
	for i, reg := range mc.regs {
		if reg.ID.Hex() == id {
			mc.regs = append(mc.regs[:i], mc.regs[i+1:]...)
			return nil
		}
	}
	return db.ErrNotFound
}

func (mc *MemDB) DeleteRegistrationByName(name string) error {
	for i, reg := range mc.regs {
		if reg.Name == name {
			mc.regs = append(mc.regs[:i], mc.regs[i+1:]...)
			return nil
		}
	}
	return db.ErrNotFound
}

func (mc *MemDB) ScrubAllRegistrations() error {
	mc.regs = make([]export.Registration, 0)
	return nil
}
