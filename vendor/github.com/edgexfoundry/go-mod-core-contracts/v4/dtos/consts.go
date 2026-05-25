//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

// JSON field name constants shared by Event and BaseReading unmarshal logic.
const (
	// shared keys
	keyApiVersion  = "apiVersion"
	keyId          = "id"
	keyDeviceName  = "deviceName"
	keyProfileName = "profileName"
	keyOrigin      = "origin"
	keyTags        = "tags"

	// Event-only keys
	keySourceName = "sourceName"
	keyReadings   = "readings"

	// BaseReading-only keys
	keyResourceName = "resourceName"
	keyValueType    = "valueType"
	keyUnits        = "units"
	keyBinaryValue  = "binaryValue"
	keyMediaType    = "mediaType"
	keyObjectValue  = "objectValue"
	keyValue        = "value"
)
