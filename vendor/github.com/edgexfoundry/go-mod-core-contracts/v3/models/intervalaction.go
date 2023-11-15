//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
)

// AuthMethod controls the authentication method to be applied to outbound http requests for interval actions
type AuthMethod string

type IntervalAction struct {
	DBTimestamp
	Id           string
	Name         string
	IntervalName string
	Content      string
	ContentType  string
	Address      Address
	AdminState   AdminState
	AuthMethod   AuthMethod
}

func (intervalAction *IntervalAction) UnmarshalJSON(b []byte) error {
	var alias struct {
		DBTimestamp
		Id           string
		Name         string
		IntervalName string
		Content      string
		ContentType  string
		Address      interface{}
		AdminState   AdminState
		AuthMethod   AuthMethod
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal intervalAction.", err)
	}
	address, err := instantiateAddress(alias.Address)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	*intervalAction = IntervalAction{
		DBTimestamp:  alias.DBTimestamp,
		Id:           alias.Id,
		Name:         alias.Name,
		IntervalName: alias.IntervalName,
		Content:      alias.Content,
		ContentType:  alias.ContentType,
		Address:      address,
		AdminState:   alias.AdminState,
		AuthMethod:   alias.AuthMethod,
	}
	return nil
}
