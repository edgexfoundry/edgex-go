//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type Transmission struct {
	Created          int64                `json:"created,omitempty"`
	Id               string               `json:"id,omitempty" validate:"omitempty,uuid"`
	Channel          Address              `json:"channel" validate:"required"`
	NotificationId   string               `json:"notificationId" validate:"required"`
	SubscriptionName string               `json:"subscriptionName" validate:"required,edgex-dto-none-empty-string"`
	Records          []TransmissionRecord `json:"records,omitempty"`
	ResendCount      int                  `json:"resendCount,omitempty"`
	Status           string               `json:"status" validate:"required,oneof='ACKNOWLEDGED' 'FAILED' 'SENT' 'ESCALATED' 'RESENDING'"`
}

// ToTransmissionModel transforms a Transmission DTO to a Transmission Model
func ToTransmissionModel(trans Transmission) models.Transmission {
	var m models.Transmission
	m.Id = trans.Id
	m.Channel = ToAddressModel(trans.Channel)
	m.Created = trans.Created
	m.NotificationId = trans.NotificationId
	m.SubscriptionName = trans.SubscriptionName
	m.Records = ToTransmissionRecordModels(trans.Records)
	m.ResendCount = trans.ResendCount
	m.Status = models.TransmissionStatus(trans.Status)
	return m
}

// ToTransmissionModels transforms a Transmission DTO array to a Transmission model array
func ToTransmissionModels(ts []Transmission) []models.Transmission {
	models := make([]models.Transmission, len(ts))
	for i, t := range ts {
		models[i] = ToTransmissionModel(t)
	}
	return models
}

// FromTransmissionModelToDTO transforms a Transmission Model to a Transmission DTO
func FromTransmissionModelToDTO(trans models.Transmission) Transmission {
	return Transmission{
		Created:          trans.Created,
		Id:               trans.Id,
		Channel:          FromAddressModelToDTO(trans.Channel),
		NotificationId:   trans.NotificationId,
		SubscriptionName: trans.SubscriptionName,
		Records:          FromTransmissionRecordModelsToDTOs(trans.Records),
		ResendCount:      trans.ResendCount,
		Status:           string(trans.Status),
	}
}

// FromTransmissionModelsToDTOs transforms a Transmission model array to a Transmission DTO array
func FromTransmissionModelsToDTOs(ts []models.Transmission) []Transmission {
	dtos := make([]Transmission, len(ts))
	for i, n := range ts {
		dtos[i] = FromTransmissionModelToDTO(n)
	}
	return dtos
}
