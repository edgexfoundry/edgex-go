/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package models

import (
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo"
)

type deviceServiceTransform interface {
	DBRefToDeviceService(dbRef mgo.DBRef) (model DeviceService, err error)
	DeviceServiceToDBRef(model DeviceService) (dbRef mgo.DBRef, err error)
}

type DeviceService struct {
	Service    `bson:",inline"`
	AdminState contract.AdminState `bson:"adminState"` // Device Service Admin State
}

func (ds *DeviceService) ToContract(transform addressableTransform) (c contract.DeviceService, err error) {
	s, err := ds.Service.ToContract(transform)
	if err != nil {
		return
	}
	c.Service = s
	c.AdminState = ds.AdminState
	return
}

func (ds *DeviceService) FromContract(from contract.DeviceService, transform addressableTransform) (id string, err error) {
	ds.AdminState = from.AdminState
	return ds.Service.FromContract(from.Service, transform)
}
