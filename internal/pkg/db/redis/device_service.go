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

	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

type redisDeviceService struct {
	models.DescribedObject
	Id             string
	Name           string
	LastConnected  int64
	LastReported   int64
	OperatingState models.OperatingState
	Addressable    string
	Labels         []string
	AdminState     models.AdminState
}

func marshalDeviceService(ds models.DeviceService) (out []byte, err error) {
	s := redisDeviceService{
		DescribedObject: ds.Service.DescribedObject,
		Id:              ds.Service.Id.Hex(),
		Name:            ds.Service.Name,
		LastConnected:   ds.Service.LastConnected,
		LastReported:    ds.Service.LastReported,
		OperatingState:  ds.Service.OperatingState,
		Addressable:     ds.Addressable.Id.Hex(),
		Labels:          ds.Labels,
		AdminState:      ds.AdminState,
	}

	return marshalObject(s)
}

func unmarshalDeviceService(o []byte, ds interface{}) (err error) {
	var s redisDeviceService

	err = unmarshalObject(o, &s)
	if err != nil {
		return err
	}

	switch x := ds.(type) {
	case *models.DeviceService:
		x.DescribedObject = s.DescribedObject
		x.Id = bson.ObjectIdHex(s.Id)
		x.Name = s.Name
		x.LastConnected = s.LastConnected
		x.LastReported = s.LastReported
		x.OperatingState = s.OperatingState
		x.Labels = s.Labels
		x.AdminState = s.AdminState
		conn, err := getConnection()
		if err != nil {
			return err
		}
		defer conn.Close()

		err = getObjectById(conn, s.Addressable, unmarshalObject, &x.Addressable)
		return err
	default:
		return fmt.Errorf("Can only unmarshal into a *DeviceService, got %T", x)
	}
}
