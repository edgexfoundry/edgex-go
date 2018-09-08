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
	"fmt"

	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

type redisScheduleEvent struct {
	models.BaseObject
	Id          string
	Name        string
	Schedule    string
	Addressable string
	Parameters  string
	Service     string
}

func marshalScheduleEvent(se models.ScheduleEvent) (out []byte, err error) {
	s := redisScheduleEvent{
		BaseObject:  se.BaseObject,
		Id:          se.Id.Hex(),
		Name:        se.Name,
		Schedule:    se.Schedule,
		Addressable: se.Addressable.Id.Hex(),
		Parameters:  se.Parameters,
		Service:     se.Service,
	}

	return marshalObject(s)
}

func unmarshalScheduleEvent(o []byte, se interface{}) (err error) {
	var s redisScheduleEvent

	err = unmarshalObject(o, &s)
	if err != nil {
		return err
	}

	switch x := se.(type) {
	case *models.ScheduleEvent:
		x.BaseObject = s.BaseObject
		x.Id = bson.ObjectIdHex(s.Id)
		x.Name = s.Name
		x.Schedule = s.Schedule
		x.Parameters = s.Parameters
		x.Service = s.Service
		conn, err := getConnection()
		if err != nil {
			return err
		}
		defer conn.Close()

		err = getObjectById(conn, s.Addressable, unmarshalObject, &x.Addressable)
		return err
	default:
		return fmt.Errorf("Can only unmarshal into a *ScheduleEvent, got %T", x)
	}
}
