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
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"net/http"
)

func getAllAddressables() ([]models.Addressable, error) {
	results := make([]models.Addressable, 0)
	err := dbClient.GetAddressables(&results)
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

func addAddressable(addressable models.Addressable) (string, error) {
	if len(addressable.Name) == 0 {
		err := errors.NewErrEmptyAddressableName()
		LoggingClient.Error(err.Error())
		return "", err
	}
	id, err := dbClient.AddAddressable(&addressable)
	if err != nil {
		if err == db.ErrNotUnique {
			err = errors.NewErrDuplicateAddressableName(addressable.Name)
		}
		LoggingClient.Error(err.Error())
		return "", err
	}

	return id.Hex(), nil // Coupling to mongo?
}

func updateAddressable(addressable models.Addressable) error {
	// Check if the addressable exists
	var res models.Addressable
	err := dbClient.GetAddressableById(&res, addressable.Id.Hex())
	if err != nil {
		if addressable.Id == "" {
			err = dbClient.GetAddressableByName(&res, addressable.Name)
		}
		if err != nil {
			if err == db.ErrNotFound {
				err = errors.NewErrAddressableNotFound(addressable.Id.Hex(), addressable.Name)
			}
			LoggingClient.Error(err.Error())
			return err
		}
	}

	// If the name is changed, check if the addressable is still in use
	if addressable.Name != "" && addressable.Name != res.Name {
		isStillInUse, err := isAddressableStillInUse(res)
		if err != nil {
			LoggingClient.Error(err.Error())
			return err
		}
		if isStillInUse {
			err = errors.NewErrAddressableInUse(res.Name)
			LoggingClient.Error(err.Error())
			return err
		}
	}

	if err := dbClient.UpdateAddressable(&addressable, &res); err != nil {
		LoggingClient.Error(err.Error())
		return err
	}

	// Notify Associates
	// TODO: Should this call be here, or in rest_addressable.go?
	if err := notifyAddressableAssociates(res, http.MethodPut); err != nil {
		LoggingClient.Error(err.Error())
		return err
	}

	return nil
}
