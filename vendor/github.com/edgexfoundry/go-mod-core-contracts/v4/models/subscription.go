//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

type Subscription struct {
	DBTimestamp
	Categories     []string
	Labels         []string
	Channels       []Address
	Description    string
	Id             string
	Receiver       string
	Name           string
	ResendLimit    int
	ResendInterval string
	AdminState     AdminState
}

// ChannelType controls the range of values which constitute valid delivery types for channels
type ChannelType string

func (subscription *Subscription) UnmarshalJSON(b []byte) error {
	var alias struct {
		DBTimestamp
		Categories     []string
		Labels         []string
		Channels       []interface{}
		Description    string
		Id             string
		Receiver       string
		Name           string
		ResendLimit    int
		ResendInterval string
		AdminState     AdminState
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal intervalAction.", err)
	}
	channels := make([]Address, len(alias.Channels))
	for i, c := range alias.Channels {
		address, err := instantiateAddress(c)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
		channels[i] = address
	}

	*subscription = Subscription{
		DBTimestamp:    alias.DBTimestamp,
		Categories:     alias.Categories,
		Labels:         alias.Labels,
		Description:    alias.Description,
		Id:             alias.Id,
		Receiver:       alias.Receiver,
		Name:           alias.Name,
		ResendLimit:    alias.ResendLimit,
		ResendInterval: alias.ResendInterval,
		Channels:       channels,
		AdminState:     alias.AdminState,
	}
	return nil
}
