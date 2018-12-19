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
package metadata

import (
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"net/http"
)

func getAllAddressables() ([]contract.Addressable, error) {
	results, err := dbClient.GetAddressables()
	if err != nil {
		LoggingClient.Error(err.Error())
		return nil, err
	}
	if len(results) > Configuration.Service.ReadMaxLimit {
		err = errors.NewErrLimitExceeded(Configuration.Service.ReadMaxLimit)
		LoggingClient.Error(err.Error())

		return nil, err
	}
	return results, nil
}

func addAddressable(addressable contract.Addressable) (string, error) {
	if len(addressable.Name) == 0 {
		err := errors.NewErrEmptyAddressableName()
		LoggingClient.Error(err.Error())
		return "", err
	}
	id, err := dbClient.AddAddressable(addressable)
	if err != nil {
		if err == db.ErrNotUnique {
			err = errors.NewErrDuplicateAddressableName(addressable.Name)
		}
		LoggingClient.Error(err.Error())
		return "", err
	}

	return id, nil // Coupling to mongo?
}

func updateAddressable(addressable contract.Addressable) error {
	var dest contract.Addressable
	var err error
	// Check if the addressable exists
	if addressable.Id == "" {
		dest, err = dbClient.GetAddressableByName(addressable.Name)
	} else {
		dest, err = dbClient.GetAddressableById(addressable.Id)
	}

	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrAddressableNotFound(addressable.Id, addressable.Name)
		}
		LoggingClient.Error(err.Error())
		return err
	}

	// If the name is changed, check if the addressable is still in use
	if addressable.Name != "" && addressable.Name != dest.Name {
		isStillInUse, err := isAddressableStillInUse(dest)
		if err != nil {
			LoggingClient.Error(err.Error())
			return err
		}
		if isStillInUse {
			err = errors.NewErrAddressableInUse(dest.Name)
			LoggingClient.Error(err.Error())
			return err
		}
	}

	dest.Name = addressable.Name
	dest.Protocol = addressable.Protocol
	dest.Address = addressable.Address
	dest.Port = addressable.Port
	dest.Path = addressable.Path
	dest.Publisher = addressable.Publisher
	dest.User = addressable.User
	dest.Password = addressable.Password
	dest.Topic = addressable.Topic

	if err := dbClient.UpdateAddressable(dest); err != nil {
		LoggingClient.Error(err.Error())
		return err
	}

	// Notify Associates
	// TODO: Should this call be here, or in rest_addressable.go?
	if err := notifyAddressableAssociates(dest, http.MethodPut); err != nil {
		LoggingClient.Error(err.Error())
		return err
	}

	return nil
}
