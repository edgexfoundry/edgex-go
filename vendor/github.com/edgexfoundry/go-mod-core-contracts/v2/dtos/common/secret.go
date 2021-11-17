//
// Copyright (c) 2020 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package common

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// SecretDataKeyValue is a key/value pair to be stored in the Secret Store as part of the Secret Data
// See detail specified by the V2 API swagger in openapi/v2
type SecretDataKeyValue struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}

// SecretRequest is the request DTO for storing supplied secret at specified Path in the Secret Store
// See detail specified by the V2 API swagger in openapi/v2
type SecretRequest struct {
	BaseRequest `json:",inline"`
	Path        string               `json:"path" validate:"required"`
	SecretData  []SecretDataKeyValue `json:"secretData" validate:"required,gt=0,dive"`
}

// Validate satisfies the Validator interface
func (sr *SecretRequest) Validate() error {
	err := common.Validate(sr)
	return err
}

// UnmarshalJSON implements the Unmarshaler interface for the SecretRequest type
func (sr *SecretRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		BaseRequest
		Path       string
		SecretData []SecretDataKeyValue
	}

	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal SecretRequest body as JSON.", err)
	}

	*sr = SecretRequest(alias)

	// validate SecretRequest DTO
	if err := sr.Validate(); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "SecretRequest validation failed.", err)
	}
	return nil
}

func NewSecretRequest(path string, secretData []SecretDataKeyValue) SecretRequest {
	return SecretRequest{
		BaseRequest: NewBaseRequest(),
		Path:        path,
		SecretData:  secretData,
	}
}
