//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type KeyData struct {
	Issuer string `json:"issuer" validate:"required"`
	Type   string `json:"type" validate:"omitempty,oneof=verification signing"`
	Key    string `json:"key" validate:"required"`
}

// ToKeyDataModel transforms the KeyData DTO to the KeyData Model
func ToKeyDataModel(keyData KeyData) models.KeyData {
	return models.KeyData{
		Issuer: keyData.Issuer,
		Type:   strings.ToLower(keyData.Type),
		Key:    keyData.Key,
	}
}
