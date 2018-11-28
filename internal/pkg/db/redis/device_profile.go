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

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/gomodule/redigo/redis"

	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

type redisDeviceProfile struct {
	models.DescribedObject
	Id              string
	Name            string
	Manufacturer    string
	Model           string
	Labels          []string
	Objects         interface{}
	DeviceResources []models.DeviceObject
	Resources       []models.ProfileResource
}

func marshalDeviceProfile(dp models.DeviceProfile) (out []byte, err error) {
	s := redisDeviceProfile{
		DescribedObject: dp.DescribedObject,
		Id:              dp.Id.Hex(),
		Name:            dp.Name,
		Manufacturer:    dp.Manufacturer,
		Model:           dp.Model,
		Labels:          dp.Labels,
		Objects:         dp.Objects,
		DeviceResources: dp.DeviceResources,
		Resources:       dp.Resources,
	}

	return marshalObject(s)
}

func unmarshalDeviceProfile(o []byte, dp interface{}) (err error) {
	var s redisDeviceProfile

	err = unmarshalObject(o, &s)
	if err != nil {
		return err
	}

	switch x := dp.(type) {
	case *models.DeviceProfile:
		x.DescribedObject = s.DescribedObject
		x.Id = bson.ObjectIdHex(s.Id)
		x.Name = s.Name
		x.Manufacturer = s.Manufacturer
		x.Model = s.Model
		x.Labels = s.Labels
		x.Objects = s.Objects
		x.DeviceResources = s.DeviceResources
		x.Resources = s.Resources
		conn, err := getConnection()
		if err != nil {
			return err
		}
		defer conn.Close()

		objects, err := getObjectsByValue(conn, db.DeviceProfile+":commands:"+s.Id)
		if err != nil {
			if err != redis.ErrNil {
				return err
			}
		}

		x.Commands = make([]models.Command, len(objects))
		for i, in := range objects {
			err = unmarshalObject(in, &x.Commands[i])
			if err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("Can only unmarshal into a *DeviceProfile, got %T", x)
	}
}
