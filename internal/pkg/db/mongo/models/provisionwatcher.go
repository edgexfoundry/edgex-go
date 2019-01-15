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

type ProvisionWatcher struct {
	Id             bson.ObjectId           `bson:"_id,omitempty"`
	Uuid           string                  `bson:"uuid,omitempty"`
	Name           string                  `bson:"name"`           // unique name and identifier of the addressable
	Identifiers    map[string]string       `bson:"identifiers"`    // set of key value pairs that identify type of of address (MAC, HTTP,...) and address to watch for (00-05-1B-A1-99-99, 10.0.0.1,...)
	Profile        mgo.DBRef               `bson:"profile"`        // device profile that should be applied to the devices available at the identifier addresses
	Service        mgo.DBRef               `bson:"service"`        // device service that owns the watcher
	OperatingState contract.OperatingState `bson:"operatingState"` // operational state - either enabled or disabled
	Created        int64                   `bson:"created"`
	Origin         int64                   `bson:"origin"`
	Modified       int64                   `bson:"modified"`
}

func (pw *ProvisionWatcher) ToContract(dpt deviceProfileTransform, dst deviceServiceTransform, ct commandTransform, at addressableTransform) (c contract.ProvisionWatcher, err error) {
	id := pw.Uuid
	if id == "" {
		id = pw.Id.Hex()
	}

	c.Id = id
	c.Name = pw.Name
	c.Identifiers = pw.Identifiers

	profile, err := dpt.DBRefToDeviceProfile(pw.Profile)
	if err != nil {
		return contract.ProvisionWatcher{}, err
	}
	c.Profile, err = profile.ToContract(ct)
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
	c.Created = pw.Created
	c.Origin = pw.Origin
	c.Modified = pw.Modified

	return
}

func (pw *ProvisionWatcher) FromContract(c contract.ProvisionWatcher, dpt deviceProfileTransform, dst deviceServiceTransform, ct commandTransform, at addressableTransform) (contractId string, err error) {
	pw.Id, pw.Uuid, err = fromContractId(c.Id)
	if err != nil {
		return
	}

	pw.Name = c.Name
	pw.Identifiers = c.Identifiers

	var profile DeviceProfile
	_, err = profile.FromContract(c.Profile, ct)
	if err != nil {
		return
	}
	pw.Profile, err = dpt.DeviceProfileToDBRef(profile)
	if err !=  nil {
		return
	}

	var service DeviceService
	err = service.FromContract(c.Service, at)
	if err != nil {
		return
	}
	pw.Service, err = dst.DeviceServiceToDBRef(service)
	if err != nil {
		return
	}

	pw.OperatingState = c.OperatingState
	pw.Created = c.Created
	pw.Origin = c.Origin
	pw.Modified = c.Modified

	return toContractId(pw.Id, pw.Uuid), nil
}
