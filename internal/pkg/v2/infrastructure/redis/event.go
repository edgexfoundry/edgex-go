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
	EventsCollection           = "cd|evt"
	EventsCollectionCreated    = EventsCollection + DBKeySeparator + v2.Created
	EventsCollectionPushed     = EventsCollection + DBKeySeparator + v2.Pushed
	EventsCollectionDeviceName = EventsCollection + DBKeySeparator + v2.Device + DBKeySeparator + v2.Name
	EventsCollectionReadings   = EventsCollection + DBKeySeparator + "readings"
)

// asyncDeleteEventsByIds deletes all events with given event Ids.  This function is implemented to be run as a separate
// goroutine in the background to achieve better performance, so this function return nothing.  When encountering any
// errors during deletion, this function will simply log the error.
func (c *Client) asyncDeleteEventsByIds(eventIds []string) {
	conn := c.Pool.Get()
	defer conn.Close()

	//start a transaction to get all events
	events, edgeXerr := getObjectsByIds(conn, common.ConvertStringsToInterfaces(eventIds))
	if edgeXerr != nil {
		c.loggingClient.Error(fmt.Sprintf("Deleted events failed while retrieving objects by Ids.  Err: %s", edgeXerr.DebugMessages()))
		return
	}

	// iterate each events for deletion in batch
	queriesInQueue := 0
	e := models.Event{}
	_ = conn.Send(MULTI)
	for i, event := range events {
		err := json.Unmarshal(event, &e)
		if err != nil {
			c.loggingClient.Error(fmt.Sprintf("unable to marshal event.  Err: %s", err.Error()))
			continue
		}
		storedKey := eventStoredKey(e.Id)
		_ = conn.Send(UNLINK, storedKey)
		_ = conn.Send(UNLINK, CreateKey(EventsCollectionReadings, e.Id))
		_ = conn.Send(ZREM, EventsCollection, storedKey)
		_ = conn.Send(ZREM, EventsCollectionCreated, storedKey)
		_ = conn.Send(ZREM, EventsCollectionPushed, storedKey)
		_ = conn.Send(ZREM, CreateKey(EventsCollectionDeviceName, e.DeviceName), storedKey)
		queriesInQueue++

		if queriesInQueue >= c.BatchSize {
			_, err = conn.Do(EXEC)
			if err != nil {
				c.loggingClient.Error(fmt.Sprintf("unable to execute batch event deletion.  Err: %s", err.Error()))
				continue
			}
			// reset queriesInQueue to zero if EXEC is successfully executed without error
			queriesInQueue = 0
			// rerun another transaction when event iteration is not finished
			if i < len(events)-1 {
				_ = conn.Send(MULTI)
			}
		}
	}

	if queriesInQueue > 0 {
		_, err := conn.Do(EXEC)
		if err != nil {
			c.loggingClient.Error(fmt.Sprintf("unable to execute batch event deletion.  Err: %s", err.Error()))
		}
	}
}

// DeletePushedEvents deletes all pushed events and corresponding readings.  This function is implemented to starts up
// two goroutines to delete readings and events in the bckground to achieve better performance.
func (c *Client) DeletePushedEvents() (edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer conn.Close()

	eventIds, readingIds, err := getEventReadingIdsByKey(conn, EventsCollectionPushed)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	c.loggingClient.Debug(fmt.Sprintf("Prepare to delete %v readings", len(readingIds)))
	go c.asyncDeleteReadingsByIds(readingIds)
	c.loggingClient.Debug(fmt.Sprintf("Prepare to delete %v events", len(eventIds)))
	go c.asyncDeleteEventsByIds(eventIds)

	return nil
}

// DeleteEventsByDeviceName deletes all pushed events and corresponding readings.  This function is implemented to starts up
// two goroutines to delete readings and events in the bckground to achieve better performance.
func (c *Client) DeleteEventsByDeviceName(deviceName string) (edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer conn.Close()

	eventIds, readingIds, err := getEventReadingIdsByKey(conn, CreateKey(EventsCollectionDeviceName, deviceName))
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	c.loggingClient.Debug(fmt.Sprintf("Prepare to delete %v readings", len(readingIds)))
	go c.asyncDeleteReadingsByIds(readingIds)
	c.loggingClient.Debug(fmt.Sprintf("Prepare to delete %v events", len(eventIds)))
	go c.asyncDeleteEventsByIds(eventIds)

	return nil
}

// ************************** DB HELPER FUNCTIONS ***************************
// eventStoredKey return the event's stored key which combines the collection name and object id
func eventStoredKey(id string) string {
	return CreateKey(EventsCollection, id)
}

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
	_ = conn.Send(ZADD, CreateKey(EventsCollectionDeviceName, e.DeviceName), e.Created, storedKey)

	// add reading ids as sorted set under each event id
	// sort by the order provided by device service
	rids := make([]interface{}, len(e.Readings)*2+1)
	rids[0] = CreateKey(EventsCollectionReadings, e.Id)
	var newReadings []models.Reading
	for i, r := range e.Readings {
		newReading, err := addReading(conn, r)
		if err != nil {
			return models.Event{}, err
		}
		newReadings = append(newReadings, newReading)

		// set the sorted set score to the index of the reading
		rids[i*2+1] = i
		rids[i*2+2] = CreateKey(ReadingsCollection, newReading.GetBaseReading().Id)
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
	_ = conn.Send(UNLINK, CreateKey(EventsCollectionReadings, e.Id))
	_ = conn.Send(ZREM, EventsCollection, storedKey)
	_ = conn.Send(ZREM, EventsCollectionCreated, storedKey)
	_ = conn.Send(ZREM, EventsCollectionPushed, storedKey)
	_ = conn.Send(ZREM, CreateKey(EventsCollectionDeviceName, e.DeviceName), storedKey)

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

func getEventReadingIdsByKey(conn redis.Conn, key string) (eventIds []string, readingIds []string, edgeXerr errors.EdgeX) {
	eventIds, err := redis.Strings(conn.Do(ZRANGEBYSCORE, key, GreaterThanZero, InfiniteMax))
	if err != nil {
		return nil, nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("retrieve event ids by key %s failed", key), err)
	}
	events, edgeXerr := getObjectsByIds(conn, common.ConvertStringsToInterfaces(eventIds))
	if edgeXerr != nil {
		return nil, nil, edgeXerr
	}
	e := models.Event{}
	for _, event := range events {
		err = json.Unmarshal(event, &e)
		if err != nil {
			return nil, nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to marshal event", err)
		}
		rIds, err := redis.Strings(conn.Do(ZRANGE, CreateKey(EventsCollectionReadings, e.Id), 0, -1))
		if err != nil {
			return nil, nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("retrieve all reading Ids of event %s failed", e.Id), err)
		}
		readingIds = append(readingIds, rIds...)
	}
	return eventIds, readingIds, nil
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
	objects, err := getObjectsByRevRange(conn, CreateKey(EventsCollectionDeviceName, name), offset, end)
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

// eventsByTimeRange query events by time range, offset, and limit
func eventsByTimeRange(conn redis.Conn, start int, end int, offset int, limit int) (events []models.Event, edgeXerr errors.EdgeX) {
	// Use following redis command to retrieve the id of events satisfied with time range/offset/limit
	// ZREVRANGEBYSCORE v2:event:created max min LIMIT offset count
	eventIds, err := redis.Strings(conn.Do(ZREVRANGEBYSCORE, EventsCollectionCreated, end, start, LIMIT, offset, limit))
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	objects, edgeXerr := getObjectsByIds(conn, common.ConvertStringsToInterfaces(eventIds))
	if edgeXerr != nil {
		return events, edgeXerr
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
