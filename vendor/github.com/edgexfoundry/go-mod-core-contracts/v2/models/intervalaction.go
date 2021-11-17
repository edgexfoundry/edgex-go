//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// IntervalAction and its properties are defined in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/support-scheduler/2.x#/IntervalAction
// Model fields are same as the DTOs documented by this swagger. Exceptions, if any, are noted below.
type IntervalAction struct {
	DBTimestamp
	Id           string
	Name         string
	IntervalName string
	Content      string
	ContentType  string
	Address      Address
	AdminState   AdminState
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
	}
	return nil
}
