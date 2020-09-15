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

const EventsCollection = "v2:event"

// ************************** DB HELPER FUNCTIONS ***************************
func addEvent(conn redis.Conn, e model.Event) (addedEvent model.Event, edgeXerr errors.EdgeX) {
	if e.Created == 0 {
		e.Created = common.MakeTimestamp()
	}

	if e.Id == "" {
		e.Id = uuid.New().String()
	}

	event := model.Event{
		CorrelationId: e.CorrelationId,
		Checksum:      e.Checksum,
		Id:            e.Id,
		Pushed:        e.Pushed,
		DeviceName:    e.DeviceName,
		Created:       e.Created,
		Origin:        e.Origin,
		Tags:          e.Tags,
	}

	m, err := json.Marshal(event)
	if err != nil {
		return addedEvent, errors.NewCommonEdgeX(errors.KindContractInvalid, "event parsing failed", err)
	}

	_ = conn.Send(MULTI)
	// use the SET command to save event as blob
	_ = conn.Send(SET, EventsCollection+":"+e.Id, m)
	_ = conn.Send(ZADD, EventsCollection, 0, e.Id)
	_ = conn.Send(ZADD, EventsCollection+":created", e.Created, e.Id)
	_ = conn.Send(ZADD, EventsCollection+":pushed", e.Pushed, e.Id)
	_ = conn.Send(ZADD, EventsCollection+":deviceName:"+e.DeviceName, e.Created, e.Id)
	if e.Checksum != "" {
		_ = conn.Send(ZADD, EventsCollection+":checksum:"+e.Checksum, 0, e.Id)
	}
	// add reading ids as sorted set under each event id
	// sort by the order provided by device service
	rids := make([]interface{}, len(e.Readings)*2+1)
	rids[0] = EventsCollection + ":readings:" + e.Id
	var newReadings []model.Reading
	for i, r := range e.Readings {
		newReading, err := addReading(conn, r)
		if err != nil {
			return model.Event{}, err
		}
		newReadings = append(newReadings, newReading)

		// set the sorted set score to the index of the reading
		rids[i*2+1] = i
		rids[i*2+2] = newReading.GetBaseReading().Id
	}
	e.Readings = newReadings
	if len(rids) > 1 {
		_ = conn.Send("ZADD", rids...)
	}

	_, err = conn.Do("EXEC")
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "event creation failed", err)
	}

	return e, edgeXerr
}
