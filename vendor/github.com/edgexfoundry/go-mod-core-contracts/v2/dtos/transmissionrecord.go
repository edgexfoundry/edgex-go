//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"encoding/json"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

type TransmissionRecord struct {
	Status   string `json:"status,omitempty" validate:"omitempty,oneof='ACKNOWLEDGED' 'FAILED' 'SENT' 'ESCALATED'"`
	Response string `json:"response,omitempty"`
	Sent     int64  `json:"sent,omitempty"`
}

// String returns a JSON encoded string representation of the object
func (tr TransmissionRecord) String() string {
	out, err := json.Marshal(tr)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

// ToTransmissionRecordModel transforms a TransmissionRecord DTO to a TransmissionRecord Model
func ToTransmissionRecordModel(tr TransmissionRecord) models.TransmissionRecord {
	var m models.TransmissionRecord
	m.Status = models.TransmissionStatus(tr.Status)
	m.Response = tr.Response
	m.Sent = tr.Sent
	return m
}

// ToTransmissionRecordModels transforms a TransmissionRecord DTO array to a TransmissionRecord model array
func ToTransmissionRecordModels(trs []TransmissionRecord) []models.TransmissionRecord {
	models := make([]models.TransmissionRecord, len(trs))
	for i, tr := range trs {
		models[i] = ToTransmissionRecordModel(tr)
	}
	return models
}

// FromTransmissionRecordModelToDTO transforms a TransmissionRecord Model to a TransmissionRecord DTO
func FromTransmissionRecordModelToDTO(tr models.TransmissionRecord) TransmissionRecord {
	return TransmissionRecord{
		Status:   string(tr.Status),
		Response: tr.Response,
		Sent:     tr.Sent,
	}
}

// FromTransmissionRecordModelsToDTOs transforms a TransmissionRecord model array to a TransmissionRecord DTO array
func FromTransmissionRecordModelsToDTOs(trs []models.TransmissionRecord) []TransmissionRecord {
	dtos := make([]TransmissionRecord, len(trs))
	for i, tr := range trs {
		dtos[i] = FromTransmissionRecordModelToDTO(tr)
	}
	return dtos
}
