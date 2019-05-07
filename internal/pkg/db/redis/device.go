/*******************************************************************************
 * Copyright 2018 Redis Labs Inc.
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
package redis

import (
	"fmt"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type redisDevice struct {
	contract.DescribedObject
	Id             string
	Name           string
	AdminState     contract.AdminState
	OperatingState contract.OperatingState
	Protocols      map[string]contract.ProtocolProperties
	AutoEvents     []contract.AutoEvent
	LastConnected  int64
	LastReported   int64
	Labels         []string
	Location       interface{}
	Service        string
	Profile        string
	ProfileName    string
}

func marshalDevice(d contract.Device) (out []byte, err error) {
	s := redisDevice{
		DescribedObject: d.DescribedObject,
		Id:              d.Id,
		Name:            d.Name,
		AdminState:      d.AdminState,
		OperatingState:  d.OperatingState,
		Protocols:       d.Protocols,
		AutoEvents:      d.AutoEvents,
		LastConnected:   d.LastConnected,
		LastReported:    d.LastReported,
		Labels:          d.Labels,
		Location:        d.Location,
		Service:         d.Service.Id,
		Profile:         d.Profile.Id,
		ProfileName:     d.Profile.Name,
	}

	return marshalObject(s)
}

func unmarshalDevice(o []byte, d interface{}) (err error) {
	var s redisDevice

	err = unmarshalObject(o, &s)
	if err != nil {
		return err
	}

	switch x := d.(type) {
	case *contract.Device:
		x.DescribedObject = s.DescribedObject
		x.Id = s.Id
		x.Name = s.Name
		x.AdminState = s.AdminState
		x.Protocols = s.Protocols
		x.AutoEvents = s.AutoEvents
		x.OperatingState = s.OperatingState
		x.LastConnected = s.LastConnected
		x.LastReported = s.LastReported
		x.Labels = s.Labels
		x.Location = s.Location

		conn, err := getConnection()
		if err != nil {
			return err
		}
		defer conn.Close()

		err = getObjectById(conn, s.Service, unmarshalDeviceService, &x.Service)
		if err != nil {
			return err
		}

		err = getObjectById(conn, s.Profile, unmarshalDeviceProfile, &x.Profile)
		if err != nil {
			return err
		}

		return nil
	default:
		return fmt.Errorf("Can only unmarshal into a *Device, got %T", x)
	}
}
