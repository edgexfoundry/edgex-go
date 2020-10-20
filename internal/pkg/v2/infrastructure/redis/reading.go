//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/pkg/common"

	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

const ReadingsCollection = "v2:reading"

var emptyBinaryValue = make([]byte, 0)

// Add a reading to the database
func addReading(conn redis.Conn, r models.Reading) (reading models.Reading, edgeXerr errors.EdgeX) {
	var m []byte
	var err error
	var baseReading *models.BaseReading
	switch newReading := r.(type) {
	case models.BinaryReading:
		// Clear the binary data since we do not want to persist binary data to save on memory.
		newReading.BinaryValue = emptyBinaryValue

		baseReading = &newReading.BaseReading
		if err = checkReadingValue(baseReading); err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
		m, err = json.Marshal(newReading)
		reading = newReading
	case models.SimpleReading:
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
	storedKey := fmt.Sprintf("%s:%s", ReadingsCollection, baseReading.Id)
	// use the SET command to save reading as blob
	_ = conn.Send(SET, storedKey, m)
	_ = conn.Send(ZADD, ReadingsCollection, 0, storedKey)
	_ = conn.Send(ZADD, fmt.Sprintf("%s:%s", ReadingsCollection, v2.Created), baseReading.Created, storedKey)
	_ = conn.Send(ZADD, fmt.Sprintf("%s:%s:%s", ReadingsCollection, v2.DeviceName, baseReading.DeviceName), baseReading.Created, storedKey)
	_ = conn.Send(ZADD, fmt.Sprintf("%s:%s:%s", ReadingsCollection, v2.Name, baseReading.Name), baseReading.Created, storedKey)

	return reading, nil
}

func checkReadingValue(b *models.BaseReading) errors.EdgeX {
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

func readingsByEventId(conn redis.Conn, eventId string) (readings []models.Reading, edgeXerr errors.EdgeX) {
	objects, err := getObjectsByRange(conn, EventsCollection+":readings:"+eventId, 0, -1)
	if errors.Kind(err) == errors.KindEntityDoesNotExist {
		return // Empty Readings in an Event is not an error
	} else if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}

	readings = make([]models.Reading, len(objects))
	for i, in := range objects {
		sr := models.SimpleReading{}
		err := json.Unmarshal(in, &sr)
		if err != nil {
			return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "reading parsing failed", err)
		}
		readings[i] = sr
	}

	return
}
