/*******************************************************************************
 * Copyright 2021 Intel Inc.
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

package secret

import (
	"encoding/json"
	"fmt"

	validation "github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/hashicorp/go-multierror"
)

// ServiceSecrets contains the list of secrets to import into a service's SecretStore
type ServiceSecrets struct {
	Secrets []ServiceSecret `json:"secrets" validate:"required,gt=0,dive"`
}

// ServiceSecret contains the information about a service's secret to import into a service's SecretStore
type ServiceSecret struct {
	Path       string                      `json:"path" validate:"edgex-dto-none-empty-string"`
	Imported   bool                        `json:"imported"`
	SecretData []common.SecretDataKeyValue `json:"secretData" validate:"required,dive"`
}

// MarshalJson marshal the service's secrets to JSON.
func (s *ServiceSecrets) MarshalJson() ([]byte, error) {
	return json.Marshal(s)
}

// UnmarshalServiceSecretsJson un-marshals the JSON containing the services list of secrets
func UnmarshalServiceSecretsJson(data []byte) (*ServiceSecrets, error) {
	secrets := &ServiceSecrets{}

	if err := json.Unmarshal(data, secrets); err != nil {
		return nil, err
	}

	if err := validation.Validate(secrets); err != nil {
		return nil, err
	}

	var validationErrs error

	// Since secretData len validation can't be specified to only validate when Imported=false, we have to do it manually here
	for _, secret := range secrets.Secrets {
		if !secret.Imported && len(secret.SecretData) == 0 {
			validationErrs = multierror.Append(validationErrs, fmt.Errorf("SecretData for '%s' must not be empty when Imported=false", secret.Path))
		}
	}

	if validationErrs != nil {
		return nil, validationErrs
	}

	return secrets, nil
}
