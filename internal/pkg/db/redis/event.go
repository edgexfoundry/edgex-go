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

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"

	correlation "github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/gomodule/redigo/redis"
)

type redisEvent struct {
	ID       string
	Checksum string
	Pushed   int64
	Device   string
	Created  int64
	Modified int64
	Origin   int64
	Tags     map[string]string
}

func marshalEvent(event correlation.Event) (out []byte, err error) {
	s := redisEvent{
		ID:       event.ID,
		Checksum: event.Checksum,
		Pushed:   event.Pushed,
		Device:   event.Device,
		Created:  event.Created,
		Modified: event.Modified,
		Origin:   event.Origin,
		Tags:     event.Tags,
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

func unmarshalEvent(o []byte) (contract.Event, error) {
	s, err := unmarshalRedisEvent(o)
	if err != nil {
		return contract.Event{}, err
	}

	event := contract.Event{
		ID:       s.ID,
		Pushed:   s.Pushed,
		Device:   s.Device,
		Created:  s.Created,
		Modified: s.Modified,
		Origin:   s.Origin,
		Tags:     s.Tags,
	}

	conn, err := getConnection()
	if err != nil {
		return contract.Event{}, err
	}
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.EventsCollection+":readings:"+s.ID, 0, -1)
	if err != nil {
		if err != redis.ErrNil {
			return contract.Event{}, err
		}
	}

	event.Readings = make([]contract.Reading, len(objects))

	for i, in := range objects {
		err = unmarshalObject(in, &event.Readings[i])
		if err != nil {
			return contract.Event{}, err
		}
	}

	return event, nil
}

// unmarshalRedisEvent constructs a redisEvent type from the provided bytes.
func unmarshalRedisEvent(o []byte) (redisEvent, error) {
	var event redisEvent

	err := json.Unmarshal(o, &event)
	if err != nil {
		return redisEvent{}, err
	}

	return event, nil
}
