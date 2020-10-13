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
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/gomodule/redigo/redis"
)

const EventsCollection = "v2:event"

// ************************** DB HELPER FUNCTIONS ***************************
func addEvent(conn redis.Conn, e models.Event) (addedEvent models.Event, edgeXerr errors.EdgeX) {
	// query Event by Id first to avoid the Id conflict
	_, edgeXerr = eventById(conn, e.Id)
	if errors.Kind(edgeXerr) != errors.KindEntityDoesNotExist {
		return addedEvent, errors.NewCommonEdgeX(errors.KindDuplicateName, "Event Id exists", nil)
	}
	edgeXerr = nil

	if e.Created == 0 {
		e.Created = common.MakeTimestamp()
	}

	event := models.Event{
		Id:         e.Id,
		Pushed:     e.Pushed,
		DeviceName: e.DeviceName,
		Created:    e.Created,
		Origin:     e.Origin,
		Tags:       e.Tags,
	}

	m, err := json.Marshal(event)
	if err != nil {
		return addedEvent, errors.NewCommonEdgeX(errors.KindContractInvalid, "event parsing failed", err)
	}

	storedKey := fmt.Sprintf("%s:%s", EventsCollection, e.Id)
	_ = conn.Send(MULTI)
	// use the SET command to save event as blob
	_ = conn.Send(SET, storedKey, m)
	_ = conn.Send(ZADD, EventsCollection, 0, storedKey)
	_ = conn.Send(ZADD, EventsCollection+":created", e.Created, storedKey)
	_ = conn.Send(ZADD, EventsCollection+":pushed", e.Pushed, storedKey)
	_ = conn.Send(ZADD, EventsCollection+":deviceName:"+e.DeviceName, e.Created, storedKey)

	// add reading ids as sorted set under each event id
	// sort by the order provided by device service
	rids := make([]interface{}, len(e.Readings)*2+1)
	rids[0] = EventsCollection + ":readings:" + e.Id
	var newReadings []models.Reading
	for i, r := range e.Readings {
		newReading, err := addReading(conn, r)
		if err != nil {
			return models.Event{}, err
		}
		newReadings = append(newReadings, newReading)

		// set the sorted set score to the index of the reading
		rids[i*2+1] = i
		rids[i*2+2] = fmt.Sprintf("%s:%s", ReadingsCollection, newReading.GetBaseReading().Id)
	}
	e.Readings = newReadings
	if len(rids) > 1 {
		_ = conn.Send(ZADD, rids...)
	}

	_, err = conn.Do(EXEC)
	if err != nil {
		edgeXerr = errors.NewCommonEdgeX(errors.KindDatabaseError, "event creation failed", err)
	}

	return e, edgeXerr
}

func eventById(conn redis.Conn, id string) (event models.Event, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectById(conn, fmt.Sprintf("%s:%s", EventsCollection, id), &event)
	if edgeXerr != nil {
		return event, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	event.Readings, edgeXerr = readingsByEventId(conn, id)
	if edgeXerr != nil {
		return event, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return
}

func (c *Client) eventTotalCount(conn redis.Conn) (uint32, errors.EdgeX) {
	count, err := redis.Int(conn.Do(ZCARD, EventsCollection))
	if err != nil {
		return 0, errors.NewCommonEdgeX(errors.KindDatabaseError, "count event failed", err)
	}

	return uint32(count), nil
}
