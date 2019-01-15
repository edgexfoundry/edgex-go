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

type DeviceProfile struct {
	DescribedObject `bson:",inline"`
	Id              bson.ObjectId     `bson:"_id,omitempty"`
	Uuid            string            `bson:"uuid,omitempty"`
	Name            string            `bson:"name"`
	Manufacturer    string            `bson:"manufacturer"`
	Model           string            `bson:"model"`
	Labels          []string          `bson:"labels"`
	Objects         interface{}       `bson:"objects"`
	DeviceResources []DeviceObject    `bson:"deviceResources"`
	Resources       []ProfileResource `bson:"resources"`
	Commands        []mgo.DBRef       `bson:"commands"`
}

func (dp *DeviceProfile) ToContract(transform commandTransform) (c contract.DeviceProfile, err error) {
	id := dp.Uuid
	if id == "" {
		id = dp.Id.Hex()
	}

	c.Id = id
	c.Name = dp.Name
	c.Manufacturer = dp.Manufacturer
	c.Model = dp.Model
	c.Labels = dp.Labels
	c.Objects = dp.Objects
	c.DescribedObject = dp.DescribedObject.ToContract()

	for _, dr := range dp.DeviceResources {
		c.DeviceResources = append(c.DeviceResources, dr.ToContract())
	}

	for _, r := range dp.Resources {
		c.Resources = append(c.Resources, r.ToContract())
	}

	for _, dbRef := range dp.Commands {
		command, err := transform.DBRefToCommand(dbRef)
		if err != nil {
			return contract.DeviceProfile{}, err
		}
		c.Commands = append(c.Commands, command.ToContract())
	}

	return
}

func (dp *DeviceProfile) FromContract(c contract.DeviceProfile, transform commandTransform) (contractId string, err error) {
	dp.Id, dp.Uuid, err = fromContractId(c.Id)
	if err != nil {
		return
	}

	dp.Name = c.Name
	dp.Manufacturer = c.Manufacturer
	dp.Model = c.Model
	dp.Labels = c.Labels
	dp.Objects = c.Objects

	dp.DescribedObject.FromContract(c.DescribedObject)

	for _, dr := range c.DeviceResources {
		var resource DeviceObject
		resource.FromContract(dr)
		dp.DeviceResources = append(dp.DeviceResources, resource)
	}

	for _, r := range c.Resources {
		var resource ProfileResource
		resource.FromContract(r)
		dp.Resources = append(dp.Resources, resource)
	}

	for _, command := range c.Commands {
		var commandModel Command
		if _, err = commandModel.FromContract(command); err != nil {
			return
		}

		var dbRef mgo.DBRef

		dbRef, err = transform.CommandToDBRef(commandModel)
		if err != nil {
			return
		}
		dp.Commands = append(dp.Commands, dbRef)
	}

	return toContractId(dp.Id, dp.Uuid), nil
}

// Custom marshaling into mongo
func (dp *DeviceProfile) GetBSON() (interface{}, error) {
	return struct {
		DescribedObject `bson:",inline"`
		Id              bson.ObjectId     `bson:"_id,omitempty"`
		Uuid            string            `bson:"uuid"`
		Name            string            `bson:"name"`         // Non-database identifier (must be unique)
		Manufacturer    string            `bson:"manufacturer"` // Manufacturer of the device
		Model           string            `bson:"model"`        // Model of the device
		Labels          []string          `bson:"labels"`       // Labels used to search for groups of profiles
		Objects         interface{}       `bson:"objects"`      // JSON data that the device service uses to communicate with devices with this profile
		DeviceResources []DeviceObject    `bson:"deviceResources"`
		Resources       []ProfileResource `bson:"resources"`
		Commands        []mgo.DBRef       `bson:"commands"` // List of commands to Get/Put information for devices associated with this profile
	}{
		DescribedObject: dp.DescribedObject,
		Id:              dp.Id,
		Uuid:			 dp.Uuid,
		Name:            dp.Name,
		Manufacturer:    dp.Manufacturer,
		Model:           dp.Model,
		Labels:          dp.Labels,
		Objects:         dp.Objects,
		DeviceResources: dp.DeviceResources,
		Resources:       dp.Resources,
		Commands:        dp.Commands,
	}, nil
}

// Custom unmarshaling out of mongo
func (dp *DeviceProfile) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		DescribedObject `bson:",inline"`
		Id              bson.ObjectId     `bson:"_id,omitempty"`
		Uuid            string            `bson:"uuid"`
		Name            string            `bson:"name"`         // Non-database identifier (must be unique)
		Manufacturer    string            `bson:"manufacturer"` // Manufacturer of the device
		Model           string            `bson:"model"`        // Model of the device
		Labels          []string          `bson:"labels"`       // Labels used to search for groups of profiles
		Objects         interface{}       `bson:"objects"`      // JSON data that the device service uses to communicate with devices with this profile
		DeviceResources []DeviceObject    `bson:"deviceResources"`
		Resources       []ProfileResource `bson:"resources"`
		Commands        []mgo.DBRef       `bson:"commands"` // List of commands to Get/Put information for devices associated with this profile
	})

	//	bsonErr := bson.Unmarshal(raw.Data, decoded)
	bsonErr := raw.Unmarshal(&decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the non-DBRef fields
	dp.DescribedObject = decoded.DescribedObject
	dp.Id = decoded.Id
	dp.Name = decoded.Name
	dp.Manufacturer = decoded.Manufacturer
	dp.Model = decoded.Model
	dp.Labels = decoded.Labels
	dp.Objects = decoded.Objects
	dp.DeviceResources = decoded.DeviceResources
	dp.Resources = decoded.Resources
	dp.Commands = decoded.Commands

	return nil
}
