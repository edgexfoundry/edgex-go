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

type redisDeviceProfile struct {
	contract.DescribedObject
	Id              string
	Name            string
	Manufacturer    string
	Model           string
	Labels          []string
	DeviceResources []contract.DeviceResource
	DeviceCommands  []contract.ProfileResource
	CoreCommands    []contract.Command
}

func marshalDeviceProfile(dp contract.DeviceProfile) (out []byte, err error) {
	s := redisDeviceProfile{
		DescribedObject: dp.DescribedObject,
		Id:              dp.Id,
		Name:            dp.Name,
		Manufacturer:    dp.Manufacturer,
		Model:           dp.Model,
		Labels:          dp.Labels,
		DeviceResources: dp.DeviceResources,
		DeviceCommands:  dp.DeviceCommands,
		CoreCommands:    dp.CoreCommands,
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
	case *contract.DeviceProfile:
		x.DescribedObject = s.DescribedObject
		x.Id = s.Id
		x.Name = s.Name
		x.Manufacturer = s.Manufacturer
		x.Model = s.Model
		x.Labels = s.Labels
		x.DeviceResources = s.DeviceResources
		x.DeviceCommands = s.DeviceCommands
		x.CoreCommands = s.CoreCommands
		return nil
	default:
		return fmt.Errorf("Can only unmarshal into a *DeviceProfile, got %T", x)
	}
}
