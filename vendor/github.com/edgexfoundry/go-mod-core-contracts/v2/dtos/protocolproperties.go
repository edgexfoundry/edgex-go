//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import "github.com/edgexfoundry/go-mod-core-contracts/v2/models"

// ProtocolProperties contains the device connection information in key/value pair
type ProtocolProperties map[string]string

// ToProtocolPropertiesModel transforms the ProtocolProperties DTO to the ProtocolProperties model
func ToProtocolPropertiesModel(p ProtocolProperties) models.ProtocolProperties {
	protocolProperties := make(models.ProtocolProperties)
	for k, v := range p {
		protocolProperties[k] = v
	}
	return protocolProperties
}

// ToProtocolModels transforms the Protocol DTO map to the Protocol model map
func ToProtocolModels(protocolDTOs map[string]ProtocolProperties) map[string]models.ProtocolProperties {
	protocolModels := make(map[string]models.ProtocolProperties)
	// Foreach the ProtocolProperties and copy values directly because the data type is map[string]string
	for k, protocolProperties := range protocolDTOs {
		protocolModels[k] = ToProtocolPropertiesModel(protocolProperties)
	}
	return protocolModels
}

// FromProtocolPropertiesModelToDTO transforms the ProtocolProperties model to the ProtocolProperties DTO
func FromProtocolPropertiesModelToDTO(p models.ProtocolProperties) ProtocolProperties {
	dto := make(ProtocolProperties)
	// Foreach the ProtocolProperties and copy values directly because the data type is map[string]string
	for k, v := range p {
		dto[k] = v
	}
	return dto
}

// FromProtocolModelsToDTOs transforms the Protocol model map to the Protocol DTO map
func FromProtocolModelsToDTOs(protocolModels map[string]models.ProtocolProperties) map[string]ProtocolProperties {
	dtos := make(map[string]ProtocolProperties)
	for k, protocolProperties := range protocolModels {
		dtos[k] = FromProtocolPropertiesModelToDTO(protocolProperties)
	}
	return dtos
}
