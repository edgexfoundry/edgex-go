/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
package mongo

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Internal version of the device profile struct
// Use this to handle DBRef
type mongoDeviceProfile struct {
	models.DeviceProfile
}

// Custom marshaling into mongo
func (mdp mongoDeviceProfile) GetBSON() (interface{}, error) {
	// Get the commands from the device profile and turn them into DBRef objects
	var dbRefs []mgo.DBRef
	for _, command := range mdp.Commands {
		dbRefs = append(dbRefs, mgo.DBRef{Collection: db.Command, Id: command.Id})
	}

	return struct {
		models.DescribedObject `bson:",inline"`
		Id                     bson.ObjectId            `bson:"_id,omitempty"`
		Name                   string                   `bson:"name"`         // Non-database identifier (must be unique)
		Manufacturer           string                   `bson:"manufacturer"` // Manufacturer of the device
		Model                  string                   `bson:"model"`        // Model of the device
		Labels                 []string                 `bson:"labels"`       // Labels used to search for groups of profiles
		Objects                interface{}              `bson:"objects"`      // JSON data that the device service uses to communicate with devices with this profile
		DeviceResources        []models.DeviceObject    `bson:"deviceResources"`
		Resources              []models.ProfileResource `bson:"resources"`
		Commands               []mgo.DBRef              `bson:"commands"` // List of commands to Get/Put information for devices associated with this profile
	}{
		DescribedObject: mdp.DescribedObject,
		Id:              mdp.Id,
		Name:            mdp.Name,
		Manufacturer:    mdp.Manufacturer,
		Model:           mdp.Model,
		Labels:          mdp.Labels,
		Objects:         mdp.Objects,
		DeviceResources: mdp.DeviceResources,
		Resources:       mdp.Resources,
		Commands:        dbRefs,
	}, nil
}

// Custom unmarshaling out of mongo
func (mdp *mongoDeviceProfile) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		models.DescribedObject `bson:",inline"`
		Id                     bson.ObjectId            `bson:"_id,omitempty"`
		Name                   string                   `bson:"name"`         // Non-database identifier (must be unique)
		Manufacturer           string                   `bson:"manufacturer"` // Manufacturer of the device
		Model                  string                   `bson:"model"`        // Model of the device
		Labels                 []string                 `bson:"labels"`       // Labels used to search for groups of profiles
		Objects                interface{}              `bson:"objects"`      // JSON data that the device service uses to communicate with devices with this profile
		DeviceResources        []models.DeviceObject    `bson:"deviceResources"`
		Resources              []models.ProfileResource `bson:"resources"`
		Commands               []mgo.DBRef              `bson:"commands"` // List of commands to Get/Put information for devices associated with this profile
	})

	//	bsonErr := bson.Unmarshal(raw.Data, decoded)
	bsonErr := raw.Unmarshal(&decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the non-DBRef fields
	mdp.DescribedObject = decoded.DescribedObject
	mdp.Id = decoded.Id
	mdp.Name = decoded.Name
	mdp.Manufacturer = decoded.Manufacturer
	mdp.Model = decoded.Model
	mdp.Labels = decoded.Labels
	mdp.Objects = decoded.Objects
	mdp.DeviceResources = decoded.DeviceResources
	mdp.Resources = decoded.Resources

	// De-reference the DBRef fields
	m, err := getCurrentMongoClient()
	if err != nil {
		return err
	}
	s := m.session.Copy()
	defer s.Close()

	comCol := s.DB(m.database.Name).C(db.Command)

	var commands []models.Command

	// Get all of the commands from the references
	for _, cRef := range decoded.Commands {
		var c models.Command
		err := comCol.FindId(cRef.Id).One(&c)
		if err != nil {
			return err
		}
		commands = append(commands, c)
	}

	mdp.Commands = commands

	return nil
}
