//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package requests

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// AddKeyDataRequest defines the Request Content for POST Key DTO.
type AddKeyDataRequest struct {
	dtoCommon.BaseRequest `json:",inline"`
	KeyData               dtos.KeyData `json:"keyData"`
}

// Validate satisfies the Validator interface
func (a *AddKeyDataRequest) Validate() error {
	err := common.Validate(a)
	return err
}

// UnmarshalJSON implements the Unmarshaler interface for the AddUserRequest type
func (a *AddKeyDataRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		dtoCommon.BaseRequest
		KeyData dtos.KeyData
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}

	*a = AddKeyDataRequest(alias)
	if err := a.Validate(); err != nil {
		return err
	}
	return nil
}
