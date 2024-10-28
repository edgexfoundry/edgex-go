//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import "github.com/edgexfoundry/go-mod-core-contracts/v4/models"

// Reading struct contains the columns of the core_data.reading table in Postgres db relates to a reading
// which includes all the fields in BaseReading, BinaryReading, SimpleReading and ObjectReading
type Reading struct {
	models.BaseReading
	EventId string `db:"event_id"` // the foreign key refers to the id column in core_data.event table
	BinaryReading
	SimpleReading
	ObjectReading
}

type SimpleReading struct {
	Value *string
}

type BinaryReading struct {
	BinaryValue []byte
	MediaType   *string
}

type ObjectReading struct {
	ObjectValue any
}

// GetBaseReading makes the Reading struct to implement the go-mod-core-contract Reading interface in models
func (r Reading) GetBaseReading() models.BaseReading {
	return models.BaseReading{
		Id:           r.Id,
		Origin:       r.Origin,
		DeviceName:   r.DeviceName,
		ResourceName: r.ResourceName,
		ProfileName:  r.ProfileName,
		ValueType:    r.ValueType,
		Units:        r.Units,
		Tags:         r.Tags,
	}
}
