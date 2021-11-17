//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// Transmission and its properties are defined in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/support-notifications/2.x#/Transmission
// Model fields are same as the DTOs documented by this swagger. Exceptions, if any, are noted below.
type Transmission struct {
	Created          int64
	Id               string
	Channel          Address
	NotificationId   string
	SubscriptionName string
	Records          []TransmissionRecord
	ResendCount      int
	Status           TransmissionStatus
}

func (trans *Transmission) UnmarshalJSON(b []byte) error {
	var alias struct {
		Created          int64
		Id               string
		Channel          interface{}
		NotificationId   string
		SubscriptionName string
		Records          []TransmissionRecord
		ResendCount      int
		Status           TransmissionStatus
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "Failed to unmarshal transmission.", err)
	}

	channel, err := instantiateAddress(alias.Channel)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	*trans = Transmission{
		Created:          alias.Created,
		Id:               alias.Id,
		Channel:          channel,
		NotificationId:   alias.NotificationId,
		SubscriptionName: alias.SubscriptionName,
		Records:          alias.Records,
		ResendCount:      alias.ResendCount,
		Status:           alias.Status,
	}
	return nil
}

// NewTransmission create transmission model with required fields
func NewTransmission(subscriptionName string, channel Address, notificationId string) Transmission {
	return Transmission{
		SubscriptionName: subscriptionName,
		Channel:          channel,
		NotificationId:   notificationId,
	}
}

// TransmissionStatus indicates the most recent success/failure of a given transmission attempt.
type TransmissionStatus string
