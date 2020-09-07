//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"

	"github.com/edgexfoundry/edgex-go/internal/pkg/common"

	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

const ReadingsCollection = "v2:reading"

var emptyBinaryValue = make([]byte, 0)

// Add a reading to the database
func addReading(conn redis.Conn, r model.Reading) (reading model.Reading, edgeXerr errors.EdgeX) {
	var m []byte
	var err error
	var baseReading *model.BaseReading
	switch newReading := r.(type) {
	case model.BinaryReading:
		// Clear the binary data since we do not want to persist binary data to save on memory.
		newReading.BinaryValue = emptyBinaryValue

		baseReading = &newReading.BaseReading
		if err = checkReadingValue(baseReading); err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
		m, err = json.Marshal(newReading)
		reading = newReading
	case model.SimpleReading:
		baseReading = &newReading.BaseReading
		if err = checkReadingValue(baseReading); err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
		m, err = json.Marshal(newReading)
		reading = newReading
	default:
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "unsupported reading type", nil)
	}

	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "reading parsing failed", err)
	}
	// use the SET command to save reading as blob
	_ = conn.Send(SET, ReadingsCollection+":"+baseReading.Id, m)
	_ = conn.Send(ZADD, ReadingsCollection, 0, baseReading.Id)
	_ = conn.Send(ZADD, ReadingsCollection+":created", baseReading.Created, baseReading.Id)
	_ = conn.Send(ZADD, ReadingsCollection+":deviceName:"+baseReading.DeviceName, baseReading.Created, baseReading.Id)
	_ = conn.Send(ZADD, ReadingsCollection+":name:"+baseReading.Name, baseReading.Created, baseReading.Id)

	return reading, nil
}

func checkReadingValue(b *model.BaseReading) errors.EdgeX {
	if b.Created == 0 {
		b.Created = common.MakeTimestamp()
	}
	// check if id is a valid uuid
	if b.Id == "" {
		b.Id = uuid.New().String()
	} else {
		_, err := uuid.Parse(b.Id)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindInvalidId, "uuid parsing failed", err)
		}
	}
	return nil
}
