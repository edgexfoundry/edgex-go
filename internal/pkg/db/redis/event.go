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
	"encoding/json"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gomodule/redigo/redis"
)

type redisEvent struct {
	ID       string
	Pushed   int64
	Device   string
	Created  int64
	Modified int64
	Origin   int64
}

func marshalEvent(event contract.Event) (out []byte, err error) {
	s := redisEvent{
		ID:       event.ID,
		Pushed:   event.Pushed,
		Device:   event.Device,
		Created:  event.Created,
		Modified: event.Modified,
		Origin:   event.Origin,
	}

	return marshalObject(s)
}

func unmarshalEvents(objects [][]byte, events []contract.Event) (err error) {
	for i, o := range objects {
		if len(o) > 0 {
			events[i], err = unmarshalEvent(o)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func unmarshalEvent(o []byte) (event contract.Event, err error) {
	var s redisEvent

	err = json.Unmarshal(o, &s)
	if err != nil {
		return contract.Event{}, err
	}

	event.ID = s.ID
	event.Pushed = s.Pushed
	event.Device = s.Device
	event.Created = s.Created
	event.Modified = s.Modified
	event.Origin = s.Origin

	conn, err := getConnection()
	if err != nil {
		return event, err
	}
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.EventsCollection+":readings:"+s.ID, 0, -1)
	if err != nil {
		if err != redis.ErrNil {
			return event, err
		}
	}

	event.Readings = make([]contract.Reading, len(objects))

	for i, in := range objects {
		err = unmarshalObject(in, &event.Readings[i])
		if err != nil {
			return event, err
		}
	}

	return event, nil
}
