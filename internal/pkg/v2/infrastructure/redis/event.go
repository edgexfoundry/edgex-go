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
)

const (
	EventsCollection           = "v2:event"
	EventsCollectionCreated    = EventsCollection + ":" + v2.Created
	EventsCollectionPushed     = EventsCollection + ":" + v2.Pushed
	EventsCollectionDeviceName = EventsCollection + ":" + v2.Device + ":" + v2.Name
	EventsCollectionReadings   = EventsCollection + ":readings"
)

// eventStoredKey return the event's stored key which combines the collection name and object id
func eventStoredKey(id string) string {
	return fmt.Sprintf("%s:%s", EventsCollection, id)
}

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

	storedKey := eventStoredKey(e.Id)
	_ = conn.Send(MULTI)
	// use the SET command to save event as blob
	_ = conn.Send(SET, storedKey, m)
	_ = conn.Send(ZADD, EventsCollection, e.Created, storedKey)
	_ = conn.Send(ZADD, EventsCollectionCreated, e.Created, storedKey)
	_ = conn.Send(ZADD, EventsCollectionPushed, e.Pushed, storedKey)
	_ = conn.Send(ZADD, fmt.Sprintf("%s:%s", EventsCollectionDeviceName, e.DeviceName), e.Created, storedKey)

	// add reading ids as sorted set under each event id
	// sort by the order provided by device service
	rids := make([]interface{}, len(e.Readings)*2+1)
	rids[0] = fmt.Sprintf("%s:%s", EventsCollectionReadings, e.Id)
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

func deleteEventById(conn redis.Conn, id string) (edgeXerr errors.EdgeX) {
	// query Event by Id first to ensure there is an corresponding event
	e, edgeXerr := eventById(conn, id)
	if edgeXerr != nil {
		return edgeXerr
	}

	// deletes all readings associated with target event
	for _, reading := range e.Readings {
		edgeXerr = deleteReadingById(conn, reading.GetBaseReading().Id)
		if edgeXerr != nil {
			return edgeXerr
		}
	}

	storedKey := eventStoredKey(e.Id)
	_ = conn.Send(MULTI)
	_ = conn.Send(UNLINK, storedKey)
	_ = conn.Send(UNLINK, fmt.Sprintf("%s:%s", EventsCollectionReadings, e.Id))
	_ = conn.Send(ZREM, EventsCollection, storedKey)
	_ = conn.Send(ZREM, EventsCollectionCreated, storedKey)
	_ = conn.Send(ZREM, EventsCollectionPushed, storedKey)
	_ = conn.Send(ZREM, fmt.Sprintf("%s:%s", EventsCollectionDeviceName, e.DeviceName), storedKey)

	res, err := redis.Values(conn.Do(EXEC))
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "event delete failed", err)
	}
	exists, _ := redis.Bool(res[0], nil)
	if !exists {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "event delete failed", redis.ErrNil)
	}

	return edgeXerr
}

func eventById(conn redis.Conn, id string) (event models.Event, edgeXerr errors.EdgeX) {
	edgeXerr = getObjectById(conn, eventStoredKey(id), &event)
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

func (c *Client) eventCountByDevice(deviceName string, conn redis.Conn) (uint32, errors.EdgeX) {
	count, err := redis.Int(conn.Do(ZCARD, fmt.Sprintf("%s:%s", EventsCollectionDeviceName, deviceName)))
	if err != nil {
		return 0, errors.NewCommonEdgeX(errors.KindDatabaseError, "count event failed", err)
	}

	return uint32(count), nil
}

func updateEventPushedById(conn redis.Conn, id string) (edgeXerr errors.EdgeX) {
	// query Event by Id first to retrieve corresponding event
	e, edgeXerr := eventById(conn, id)
	if edgeXerr != nil {
		return edgeXerr
	}

	// update the pushed timestamp
	event := models.Event{
		Id:         e.Id,
		Pushed:     common.MakeTimestamp(),
		DeviceName: e.DeviceName,
		Created:    e.Created,
		Origin:     e.Origin,
		Tags:       e.Tags,
	}
	m, err := json.Marshal(event)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "event parsing failed", err)
	}

	storedKey := eventStoredKey(event.Id)
	_ = conn.Send(MULTI)
	// use the SET command to overwrite the updated event as blob
	_ = conn.Send(SET, storedKey, m)
	// EventsCollectionPushed sorted set uses pushed as the score, and ZADD command will update the score of the event
	_ = conn.Send(ZADD, EventsCollectionPushed, event.Pushed, storedKey)

	_, err = conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, "event pushed update failed", err)
	}

	return nil
}

func (c *Client) allEvents(conn redis.Conn, offset int, limit int) (events []models.Event, edgeXerr errors.EdgeX) {
	end := offset + limit - 1
	if limit == -1 { //-1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	objects, err := getObjectsByRevRange(conn, EventsCollection, offset, end)
	if err != nil {
		return events, errors.NewCommonEdgeXWrapper(err)
	}

	events = make([]models.Event, len(objects))
	for i, in := range objects {
		e := models.Event{}
		err := json.Unmarshal(in, &e)
		if err != nil {
			return []models.Event{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "event format parsing failed from the database", err)
		}
		e.Readings, edgeXerr = readingsByEventId(conn, e.Id)
		if edgeXerr != nil {
			return events, errors.NewCommonEdgeXWrapper(edgeXerr)
		}
		events[i] = e
	}
	return events, nil
}

// eventsByDeviceName query events by offset, limit and device name
func eventsByDeviceName(conn redis.Conn, offset int, limit int, name string) (events []models.Event, edgeXerr errors.EdgeX) {
	end := offset + limit - 1
	if limit == -1 { //-1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
		end = limit
	}
	objects, err := getObjectsByRevRange(conn, fmt.Sprintf("%s:%s", EventsCollectionDeviceName, name), offset, end)
	if err != nil {
		return events, errors.NewCommonEdgeXWrapper(err)
	}

	events = make([]models.Event, len(objects))
	for i, in := range objects {
		e := models.Event{}
		err := json.Unmarshal(in, &e)
		if err != nil {
			return events, errors.NewCommonEdgeX(errors.KindContractInvalid, "event parsing failed", err)
		}
		e.Readings, edgeXerr = readingsByEventId(conn, e.Id)
		if edgeXerr != nil {
			return events, errors.NewCommonEdgeXWrapper(edgeXerr)
		}
		events[i] = e
	}
	return events, nil
}
