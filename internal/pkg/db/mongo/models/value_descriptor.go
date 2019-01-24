/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
	"github.com/globalsign/mgo/bson"
)

type ValueDescriptor struct {
	Id           bson.ObjectId `bson:"_id,omitempty"`
	Uuid         string        `bson:"uuid"`
	Created      int64         `bson:"created"`
	Description  string        `bson:"description,omitempty"`
	Modified     int64         `bson:"modified"`
	Origin       int64         `bson:"origin"`
	Name         string        `bson:"name"`
	Min          interface{}   `bson:"min,omitempty"`
	Max          interface{}   `bson:"max,omitempty"`
	DefaultValue interface{}   `bson:"defaultValue,omitempty"`
	Type         string        `bson:"type,omitempty"`
	UomLabel     string        `bson:"uomLabel,omitempty"`
	Formatting   string        `bson:"formatting,omitempty"`
	Labels       []string      `bson:"labels,omitempty"`
}

func (v *ValueDescriptor) ToContract() contract.ValueDescriptor {
	// Always hand back the UUID as the contract event ID unless it's blank (an old event, for example blackbox test scripts)
	id := v.Uuid
	if id == "" {
		id = v.Id.Hex()
	}
	to := contract.ValueDescriptor{
		Id:           id,
		Created:      v.Created,
		Description:  v.Description,
		Modified:     v.Modified,
		Origin:       v.Origin,
		Name:         v.Name,
		Min:          v.Min,
		Max:          v.Max,
		DefaultValue: v.DefaultValue,
		Type:         v.Type,
		UomLabel:     v.UomLabel,
		Formatting:   v.Formatting,
		Labels:       []string{},
	}
	for _, l := range v.Labels {
		to.Labels = append(to.Labels, l)
	}
	return to
}

func (v *ValueDescriptor) FromContract(from contract.ValueDescriptor) (id string, err error) {
	v.Id, v.Uuid, err = fromContractId(from.Id)
	if err != nil {
		return
	}

	v.Created = from.Created
	v.Description = from.Description
	v.Modified = from.Modified
	v.Origin = from.Origin
	v.Name = from.Name
	v.Min = from.Min
	v.Max = from.Max
	v.DefaultValue = from.DefaultValue
	v.Type = from.Type
	v.UomLabel = from.UomLabel
	v.Formatting = from.Formatting
	v.Labels = []string{}

	for _, l := range from.Labels {
		v.Labels = append(v.Labels, l)
	}

	if v.Created == 0 {
		v.Created = db.MakeTimestamp()
	}

	id = toContractId(v.Id, v.Uuid)
	return
}
