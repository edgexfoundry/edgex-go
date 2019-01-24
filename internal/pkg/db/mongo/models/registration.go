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
	"github.com/globalsign/mgo/bson"
)

type Filter struct {
	DeviceIDs          []string `bson:"deviceIdentifiers,omitempty"`
	ValueDescriptorIDs []string `bson:"valueDescriptorIdentifiers,omitempty"`
}

type EncryptionDetails struct {
	Algo       string `bson:"encryptionAlgorithm,omitempty"`
	Key        string `bson:"encryptionKey,omitempty"`
	InitVector string `bson:"initializingVector,omitempty"`
}

type Registration struct {
	Created     int64         `bson:"created"`
	Modified    int64         `bson:"modified"`
	Origin      int64         `bson:"origin"`
	ID          bson.ObjectId `bson:"_id,omitempty"`
	Uuid        string        `bson:"uuid,omitempty"`
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

	c.Filter.DeviceIDs = r.Filter.DeviceIDs
	c.Filter.ValueDescriptorIDs = r.Filter.ValueDescriptorIDs

	c.Encryption.Algo = r.Encryption.Algo
	c.Encryption.Key = r.Encryption.Key
	c.Encryption.InitVector = r.Encryption.InitVector

	c.Compression = r.Compression
	c.Enable = r.Enable
	c.Destination = r.Destination

	return
}

func (r *Registration) FromContract(from contract.Registration) (id string, err error) {
	r.ID, r.Uuid, err = fromContractId(from.ID)
	if err != nil {
		return
	}

	r.Created = from.Created
	r.Modified = from.Modified
	r.Origin = from.Origin
	r.Name = from.Name

	r.Addressable = Addressable{}
	_, err = r.Addressable.FromContract(from.Addressable)
	if err != nil {
		return
	}

	r.Format = from.Format

	r.Filter.DeviceIDs = from.Filter.DeviceIDs
	r.Filter.ValueDescriptorIDs = from.Filter.ValueDescriptorIDs

	r.Encryption.Algo = from.Encryption.Algo
	r.Encryption.Key = from.Encryption.Key
	r.Encryption.InitVector = from.Encryption.InitVector

	r.Compression = from.Compression
	r.Enable = from.Enable
	r.Destination = from.Destination

	id = toContractId(r.ID, r.Uuid)
	return
}

func (r *Registration) TimestampForUpdate() {
	r.Modified = db.MakeTimestamp()
}

func (r *Registration) TimestampForAdd() {
	r.TimestampForUpdate()
	r.Created = r.Modified
}
