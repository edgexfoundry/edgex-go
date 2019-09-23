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
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type ProvisionWatcher struct {
	Created             int64                   `bson:"created"`
	Modified            int64                   `bson:"modified"`
	Origin              int64                   `bson:"origin"`
	Id                  bson.ObjectId           `bson:"_id,omitempty"`
	Uuid                string                  `bson:"uuid,omitempty"`
	Name                string                  `bson:"name"`                // unique name and identifier of the addressable
	Identifiers         map[string]string       `bson:"identifiers"`         // set of key value pairs that identify type of of address (MAC, HTTP,...) and address to watch for (00-05-1B-A1-99-99, 10.0.0.1,...)
	BlockingIdentifiers map[string][]string     `bson:"blockingidentifiers"` // set of keys and blocking values that disallow matches made on Identifiers
	Profile             mgo.DBRef               `bson:"profile"`             // device profile that should be applied to the devices available at the identifier addresses
	Service             mgo.DBRef               `bson:"service"`             // device service that owns the watcher
	OperatingState      contract.OperatingState `bson:"operatingState"`      // operational state - either enabled or disabled
	AdminState          contract.AdminState     `bson:"adminState"`          // administrative state - either unlocked or locked
}

func (pw *ProvisionWatcher) ToContract(dpt deviceProfileTransform, dst deviceServiceTransform, at addressableTransform) (c contract.ProvisionWatcher, err error) {
	id := pw.Uuid
	if id == "" {
		id = pw.Id.Hex()
	}

	c.Id = id
	c.Created = pw.Created
	c.Modified = pw.Modified
	c.Origin = pw.Origin
	c.Name = pw.Name
	c.Identifiers = pw.Identifiers
	c.BlockingIdentifiers = pw.BlockingIdentifiers

	profile, err := dpt.DBRefToDeviceProfile(pw.Profile)
	if err != nil {
		return contract.ProvisionWatcher{}, err
	}
	c.Profile, err = profile.ToContract()
	if err != nil {
		return contract.ProvisionWatcher{}, err
	}

	service, err := dst.DBRefToDeviceService(pw.Service)
	if err != nil {
		return contract.ProvisionWatcher{}, err
	}
	c.Service, err = service.ToContract(at)
	if err != nil {
		return contract.ProvisionWatcher{}, err
	}

	c.OperatingState = pw.OperatingState
	c.AdminState = pw.AdminState

	return
}

func (pw *ProvisionWatcher) FromContract(from contract.ProvisionWatcher, dpt deviceProfileTransform, dst deviceServiceTransform, at addressableTransform) (id string, err error) {
	pw.Id, pw.Uuid, err = fromContractId(from.Id)
	if err != nil {
		return
	}

	pw.Created = from.Created
	pw.Modified = from.Modified
	pw.Origin = from.Origin
	pw.Name = from.Name
	pw.Identifiers = from.Identifiers
	pw.BlockingIdentifiers = from.BlockingIdentifiers

	var profile DeviceProfile
	if _, err = profile.FromContract(from.Profile); err != nil {
		return
	}
	if pw.Profile, err = dpt.DeviceProfileToDBRef(profile); err != nil {
		return
	}

	var service DeviceService
	if _, err = service.FromContract(from.Service, at); err != nil {
		return
	}
	if pw.Service, err = dst.DeviceServiceToDBRef(service); err != nil {
		return
	}

	pw.OperatingState = from.OperatingState
	pw.AdminState = from.AdminState

	id = toContractId(pw.Id, pw.Uuid)
	return
}

func (pw *ProvisionWatcher) TimestampForUpdate() {
	pw.Modified = db.MakeTimestamp()
}

func (pw *ProvisionWatcher) TimestampForAdd() {
	pw.TimestampForUpdate()
	pw.Created = pw.Modified
}
