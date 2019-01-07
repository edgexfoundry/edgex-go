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
)

type DeviceService struct {
	Service    `bson:",inline"`
	AdminState contract.AdminState `bson:"adminState"` // Device Service Admin State
}

func (ds *DeviceService) ToContract() contract.DeviceService {
	return contract.DeviceService{
		Service:    ds.Service.ToContract(),
		AdminState: ds.AdminState,
	}
}

func (ds *DeviceService) FromContract(from contract.DeviceService) error {
	var err error
	if err = ds.Service.FromContract(from.Service); err != nil {
		return err
	}
	ds.AdminState = from.AdminState
	return nil
}
