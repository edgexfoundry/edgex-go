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
	"github.com/globalsign/mgo/bson"
)

type Registration struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	Uuid        string        `bson:"uuid,omitempty"`
	Created     int64
	Modified    int64
	Origin      int64
	Name        string
	Addressable Addressable
	Format      string
	Filter      Filter
	Encryption  EncryptionDetails
	Compression string
	Enable      bool
	Destination string
}

func (r *Registration) ToContract() (c contract.Registration) {
	id := r.Uuid
	if id == "" {
		id = r.ID.Hex()
	}

	c.ID = id
	c.Created = r.Created
	c.Modified = r.Modified
	c.Origin = r.Origin
	c.Name = r.Name
	c.Addressable = r.Addressable.ToContract()
	c.Format = r.Format
	c.Filter = r.Filter.ToContract()
	c.Encryption = r.Encryption.ToContract()
	c.Compression = r.Compression
	c.Enable = r.Enable
	c.Destination = r.Destination

	return
}

func (r *Registration) FromContract(c contract.Registration) (contractId string, err error){
	r.ID, r.Uuid, err = fromContractId(c.ID)
	if err != nil {
		return
	}

	r.Created = c.Created
	r.Modified = c.Modified
	r.Origin = c.Origin
	r.Name = c.Name

	r.Addressable = Addressable{}
	err = r.Addressable.FromContract(c.Addressable)
	if err != nil {
		return
	}

	r.Format = c.Format

	r.Filter = Filter{}
	r.Filter.FromContract(c.Filter)

	r.Encryption = EncryptionDetails{}
	r.Encryption.FromContract(c.Encryption)

	r.Compression = c.Compression
	r.Enable = c.Enable
	r.Destination = c.Destination

	return toContractId(r.ID, r.Uuid), nil
}
