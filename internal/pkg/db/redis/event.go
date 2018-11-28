/*******************************************************************************
 * Copyright 2018 Redis Labs Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package redis

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gomodule/redigo/redis"
	"gopkg.in/mgo.v2/bson"
)

type redisEvent struct {
	ID       string
	Pushed   int64
	Device   string
	Created  int64
	Modified int64
	Origin   int64
	Event    string
}

func marshalEvent(event models.Event) (out []byte, err error) {
	s := redisEvent{
		ID:       event.ID.Hex(),
		Pushed:   event.Pushed,
		Device:   event.Device,
		Created:  event.Created,
		Modified: event.Modified,
		Origin:   event.Origin,
		Event:    event.Event,
	}

	return marshalObject(s)
}

func unmarshalEvents(objects [][]byte, events []models.Event) (err error) {
	for i, o := range objects {
		if len(o) > 0 {
			err := unmarshalEvent(o, &events[i])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func unmarshalEvent(o []byte, event *models.Event) (err error) {
	var s redisEvent

	err = bson.Unmarshal(o, &s)
	if err != nil {
		return err
	}

	event.ID = bson.ObjectIdHex(s.ID)
	event.Pushed = s.Pushed
	event.Device = s.Device
	event.Created = s.Created
	event.Modified = s.Modified
	event.Origin = s.Origin
	event.Event = s.Event

	conn, err := getConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.EventsCollection+":readings:"+s.ID, 0, -1)
	if err != nil {
		if err != redis.ErrNil {
			return err
		}
	}

	event.Readings = make([]models.Reading, len(objects))

	for i, in := range objects {
		err = unmarshalObject(in, &event.Readings[i])
		if err != nil {
			return err
		}
	}

	return nil
}
