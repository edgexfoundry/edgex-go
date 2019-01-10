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
	"github.com/globalsign/mgo/bson"
)

type Service struct {
	DescribedObject `bson:",inline"`
	Id              bson.ObjectId           `bson:"_id,omitempty"`
	Uuid            string                  `bson:"uuid,omitempty"`
	Name            string                  `bson:"name"`           // time in milliseconds that the device last provided any feedback or responded to any request
	LastConnected   int64                   `bson:"lastConnected"`  // time in milliseconds that the device last reported data to the core
	LastReported    int64                   `bson:"lastReported"`   // operational state - either enabled or disabled
	OperatingState  contract.OperatingState `bson:"operatingState"` // operational state - ether enabled or disableddc
	Labels          []string                `bson:"labels"`         // tags or other labels applied to the device service for search or other identification needs
	Addressable     mgo.DBRef               `bson:"addressable"`    // address (MQTT topic, HTTP address, serial bus, etc.) for reaching the service
}

func (s *Service) ToContract(transform addressableTransform) (c contract.Service, err error) {
	// Always hand back the UUID as the contract command ID unless it's blank (an old command, for example blackbox test scripts)
	id := s.Uuid
	if id == "" {
		id = s.Id.Hex()
	}

	c.DescribedObject = s.DescribedObject.ToContract()
	c.Id = id
	c.Name = s.Name
	c.LastConnected = s.LastConnected
	c.LastReported = s.LastReported
	c.OperatingState = s.OperatingState
	c.Labels = s.Labels

	a, err := transform.DBRefToAddressable(s.Addressable)
	c.Addressable = a.ToContract()
	return
}

func (s *Service) FromContract(from contract.Service, transform addressableTransform) (err error) {
	s.Id, s.Uuid, err = fromContractId(from.Id)
	if err != nil {
		return err
	}

	s.DescribedObject.FromContract(from.DescribedObject)
	s.Name = from.Name
	s.LastConnected = from.LastConnected
	s.LastReported = from.LastReported
	s.OperatingState = from.OperatingState
	s.Labels = from.Labels

	var addrModel Addressable
	err = addrModel.FromContract(from.Addressable)
	if err != nil {
		return
	}
	s.Addressable, err = transform.AddressableToDBRef(addrModel)
	return
}
