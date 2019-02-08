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
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type deviceProfileTransform interface {
	DBRefToDeviceProfile(dbRef mgo.DBRef) (model DeviceProfile, err error)
	DeviceProfileToDBRef(model DeviceProfile) (dbRef mgo.DBRef, err error)
}

type PropertyValue struct {
	Type         string `bson:"type"`         // ValueDescriptor Type of property after transformations
	ReadWrite    string `bson:"readWrite"`    // Read/Write Permissions set for this property
	Minimum      string `bson:"minimum"`      // Minimum value that can be get/set from this property
	Maximum      string `bson:"maximum"`      // Maximum value that can be get/set from this property
	DefaultValue string `bson:"defaultValue"` // Default value set to this property if no argument is passed
	Size         string `bson:"size"`         // Size of this property in its type  (i.e. bytes for numeric types, characters for string types)
	Mask         string `bson:"mask"`         // Mask to be applied prior to get/set of property
	Shift        string `bson:"shift"`        // Shift to be applied after masking, prior to get/set of property
	Scale        string `bson:"scale"`        // Multiplicative factor to be applied after shifting, prior to get/set of property
	Offset       string `bson:"offset"`       // Additive factor to be applied after multiplying, prior to get/set of property
	Base         string `bson:"base"`         // Base for property to be applied to, leave 0 for no power operation (i.e. base ^ property: 2 ^ 10)
	Assertion    string `bson:"assertion"`    // Required value of the property, set for checking error state.  Failing an assertion condition will mark the device with an error state
	Precision    string `bson:"precision"`
}

type Units struct {
	Type         string `bson:"type"`
	ReadWrite    string `bson:"readWrite"`
	DefaultValue string `bson:"defaultValue"`
}

type ProfileProperty struct {
	Value PropertyValue `bson:"value"`
	Units Units         `bson:"units"`
}

type DeviceResource struct {
	Description string                 `bson:"description"`
	Name        string                 `bson:"name"`
	Tag         string                 `bson:"tag"`
	Properties  ProfileProperty        `bson:"properties"`
	Attributes  map[string]interface{} `bson:"attributes"`
}

type ResourceOperation struct {
	Index     string            `bson:"index"`
	Operation string            `bson:"operation"`
	Object    string            `bson:"object"`
	Parameter string            `bson:"parameter"`
	Resource  string            `bson:"resource"`
	Secondary []string          `bson:"secondary"`
	Mappings  map[string]string `bson:"mappings"`
}

type ProfileResource struct {
	Name string              `bson:"name"`
	Get  []ResourceOperation `bson:"get"`
	Set  []ResourceOperation `bson:"set"`
}

type DeviceProfile struct {
	Created         int64             `bson:"created"`
	Modified        int64             `bson:"modified"`
	Origin          int64             `bson:"origin"`
	Description     string            `bson:"description"`
	Id              bson.ObjectId     `bson:"_id,omitempty"`
	Uuid            string            `bson:"uuid,omitempty"`
	Name            string            `bson:"name"`
	Manufacturer    string            `bson:"manufacturer"`
	Model           string            `bson:"model"`
	Labels          []string          `bson:"labels"`
	DeviceResources []DeviceResource  `bson:"deviceResources"`
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
	c.Created = dp.Created
	c.Modified = dp.Modified
	c.Origin = dp.Origin
	c.Description = dp.Description

	for _, dr := range dp.DeviceResources {
		var cdo contract.DeviceResource

		cdo.Description = dr.Description
		cdo.Name = dr.Name
		cdo.Tag = dr.Tag

		cdo.Properties.Value.Type = dr.Properties.Value.Type
		cdo.Properties.Value.ReadWrite = dr.Properties.Value.ReadWrite
		cdo.Properties.Value.Minimum = dr.Properties.Value.Minimum
		cdo.Properties.Value.Maximum = dr.Properties.Value.Maximum
		cdo.Properties.Value.DefaultValue = dr.Properties.Value.DefaultValue
		cdo.Properties.Value.Size = dr.Properties.Value.Size
		cdo.Properties.Value.Mask = dr.Properties.Value.Mask
		cdo.Properties.Value.Shift = dr.Properties.Value.Shift
		cdo.Properties.Value.Scale = dr.Properties.Value.Scale
		cdo.Properties.Value.Offset = dr.Properties.Value.Offset
		cdo.Properties.Value.Base = dr.Properties.Value.Base
		cdo.Properties.Value.Assertion = dr.Properties.Value.Assertion
		cdo.Properties.Value.Precision = dr.Properties.Value.Precision

		cdo.Properties.Units.Type = dr.Properties.Units.Type
		cdo.Properties.Units.ReadWrite = dr.Properties.Units.ReadWrite
		cdo.Properties.Units.DefaultValue = dr.Properties.Units.DefaultValue

		cdo.Attributes = dr.Attributes

		c.DeviceResources = append(c.DeviceResources, cdo)
	}

	for _, r := range dp.Resources {
		var cpr contract.ProfileResource
		cpr.Name = r.Name
		for _, ro := range r.Get {
			cpr.Get = append(cpr.Get, contract.ResourceOperation{
				Index:     ro.Index,
				Operation: ro.Operation,
				Object:    ro.Object,
				Parameter: ro.Parameter,
				Resource:  ro.Resource,
				Secondary: ro.Secondary,
				Mappings:  ro.Mappings,
			})
		}

		for _, ro := range r.Set {
			cpr.Set = append(cpr.Set, contract.ResourceOperation{
				Index:     ro.Index,
				Operation: ro.Operation,
				Object:    ro.Object,
				Parameter: ro.Parameter,
				Resource:  ro.Resource,
				Secondary: ro.Secondary,
				Mappings:  ro.Mappings,
			})
		}

		c.Resources = append(c.Resources, cpr)
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

func (dp *DeviceProfile) FromContract(from contract.DeviceProfile, transform commandTransform) (contractId string, err error) {
	dp.Id, dp.Uuid, err = fromContractId(from.Id)
	if err != nil {
		return
	}

	dp.Name = from.Name
	dp.Manufacturer = from.Manufacturer
	dp.Model = from.Model
	dp.Labels = from.Labels

	dp.Created = from.Created
	dp.Modified = from.Modified
	dp.Origin = from.Origin
	dp.Description = from.Description

	for _, dr := range from.DeviceResources {
		var do DeviceResource

		do.Description = dr.Description
		do.Name = dr.Name
		do.Tag = dr.Tag

		do.Properties.Value.Type = dr.Properties.Value.Type
		do.Properties.Value.ReadWrite = dr.Properties.Value.ReadWrite
		do.Properties.Value.Minimum = dr.Properties.Value.Minimum
		do.Properties.Value.Maximum = dr.Properties.Value.Maximum
		do.Properties.Value.DefaultValue = dr.Properties.Value.DefaultValue
		do.Properties.Value.Size = dr.Properties.Value.Size
		do.Properties.Value.Mask = dr.Properties.Value.Mask
		do.Properties.Value.Shift = dr.Properties.Value.Shift
		do.Properties.Value.Scale = dr.Properties.Value.Scale
		do.Properties.Value.Offset = dr.Properties.Value.Offset
		do.Properties.Value.Base = dr.Properties.Value.Base
		do.Properties.Value.Assertion = dr.Properties.Value.Assertion
		do.Properties.Value.Precision = dr.Properties.Value.Precision

		do.Properties.Units.Type = dr.Properties.Units.Type
		do.Properties.Units.ReadWrite = dr.Properties.Units.ReadWrite
		do.Properties.Units.DefaultValue = dr.Properties.Units.DefaultValue

		do.Attributes = dr.Attributes

		dp.DeviceResources = append(dp.DeviceResources, do)
	}

	for _, r := range from.Resources {
		var pr ProfileResource
		pr.Name = r.Name
		for _, ro := range r.Get {
			pr.Get = append(pr.Get, ResourceOperation{
				Index:     ro.Index,
				Operation: ro.Operation,
				Object:    ro.Object,
				Parameter: ro.Parameter,
				Resource:  ro.Resource,
				Secondary: ro.Secondary,
				Mappings:  ro.Mappings,
			})
		}

		for _, ro := range r.Set {
			pr.Set = append(pr.Set, ResourceOperation{
				Index:     ro.Index,
				Operation: ro.Operation,
				Object:    ro.Object,
				Parameter: ro.Parameter,
				Resource:  ro.Resource,
				Secondary: ro.Secondary,
				Mappings:  ro.Mappings,
			})
		}

		dp.Resources = append(dp.Resources, pr)
	}

	for _, command := range from.Commands {
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

	contractId = toContractId(dp.Id, dp.Uuid)
	return
}

func (dp *DeviceProfile) TimestampForUpdate() {
	dp.Modified = db.MakeTimestamp()
}

func (dp *DeviceProfile) TimestampForAdd() {
	dp.TimestampForUpdate()
	dp.Created = dp.Modified
}
