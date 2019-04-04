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

type redisProvisionWatcher struct {
	contract.BaseObject
	Id             string
	Name           string
	Identifiers    map[string]string
	Profile        string
	Service        string
	OperatingState contract.OperatingState
}

func marshalProvisionWatcher(pw contract.ProvisionWatcher) (out []byte, err error) {
	s := redisProvisionWatcher{
		BaseObject:     pw.BaseObject,
		Id:             pw.Id,
		Name:           pw.Name,
		Identifiers:    pw.Identifiers,
		Profile:        pw.Profile.Id,
		Service:        pw.Service.Id,
		OperatingState: pw.OperatingState,
	}

	return marshalObject(s)
}

func unmarshalProvisionWatcher(o []byte, pw interface{}) (err error) {
	var s redisProvisionWatcher

	err = unmarshalObject(o, &s)
	if err != nil {
		return err
	}

	switch x := pw.(type) {
	case *contract.ProvisionWatcher:
		x.BaseObject = s.BaseObject
		x.Id = s.Id
		x.Name = s.Name
		x.Identifiers = s.Identifiers
		x.OperatingState = s.OperatingState

		conn, err := getConnection()
		if err != nil {
			return err
		}
		defer conn.Close()

		err = getObjectById(conn, s.Profile, unmarshalDeviceProfile, &x.Profile)
		if err != nil {
			return err
		}

		err = getObjectById(conn, s.Service, unmarshalDeviceService, &x.Service)
		if err != nil {
			return err
		}

		return nil
	default:
		return fmt.Errorf("Can only unmarshal into a *ProvisionWatcher, got %T", x)
	}
}
