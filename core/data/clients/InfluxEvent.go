/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
 *
 * @microservice: core-data-go library
 * @author: Ryan Comer, Dell & Masataka Mizukoshi
 * @version: 0.5.0
 *******************************************************************************/

package clients

import (
	"fmt"
	"strconv"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/influxdata/influxdb/client/v2"
)

// Struct that wraps an event to handle DBRefs
type InfluxEvent struct {
	models.Event
}

// Custom marshaling into influxdb
func (ie InfluxEvent) Get() (interface{}, error) {
	// Turn the readings into DBRef objects
	var readings []mgo.DBRef
	for _, reading := range ie.Readings {
		readings = append(readings, mgo.DBRef{Collection: READINGS_COLLECTION, Id: reading.Id})
	}

	return struct {
		ID       bson.ObjectId `json:"_id,omitempty"`
		Pushed   int64         `json:"pushed"`
		Device   string        `json:"device"` // Device identifier (name or id)
		Created  int64         `json:"created"`
		Modified int64         `json:"modified"`
		Origin   int64         `json:"origin"`
		Schedule string        `json:"schedule,omitempty"` // Schedule identifier
		Event    string        `json:"event"`              // Schedule event identifier
		Readings []mgo.DBRef   `json:"readings"`           // List of readings
	}{
		ID:       ie.ID,
		Pushed:   ie.Pushed,
		Device:   ie.Device,
		Created:  ie.Created,
		Modified: ie.Modified,
		Origin:   ie.Origin,
		Schedule: ie.Schedule,
		Event:    ie.Event.Event,
		Readings: readings,
	}, nil
}

// Custom unmarshaling out of influxdb
func (ie *InfluxEvent) Set(raw bson.Raw) error {
	decoded := new(struct {
		ID       bson.ObjectId `json:"_id,omitempty"`
		Pushed   int64         `json:"pushed"`
		Device   string        `json:"device"` // Device identifier (name or id)
		Created  int64         `json:"created"`
		Modified int64         `json:"modified"`
		Origin   int64         `json:"origin"`
		Schedule string        `json:"schedule,omitempty"` // Schedule identifier
		Event    string        `json:"event"`              // Schedule event identifier
		Readings []mgo.DBRef   `json:"readings"`           // List of readings
	})

	bsonErr := raw.Unmarshal(decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the non-DBRef fields
	ie.ID = decoded.ID
	ie.Pushed = decoded.Pushed
	ie.Device = decoded.Device
	ie.Created = decoded.Created
	ie.Modified = decoded.Modified
	ie.Origin = decoded.Origin
	ie.Schedule = decoded.Schedule
	ie.Event.Event = decoded.Event

	// De-reference the DBRef fields
	ic, err := getCurrentInfluxClient()
	if err != nil {
		return err
	}

	var readings []models.Reading

	// Get all of the reading objects
	for _, rRef := range decoded.Readings {
		var reading models.Reading
		q := fmt.Sprintf("SELECT * FROM %s WHERE id = '%s'", READINGS_COLLECTION, rRef.Id)
		res, err := queryDB(ic.Client, q, ic.Database)
		if err != nil {
			return err
		}
		reading.Id = rRef.Id.(bson.ObjectId)
		for i, col := range res[0].Series[0].Columns {
		switch col {
			case "pushed":
				n, err := strconv.ParseInt(res[0].Series[0].Values[0][i].(string), 10, 64)
				if err != nil {
					return err
				}
				reading.Pushed = n	
			case "created":
				n, err := strconv.ParseInt(res[0].Series[0].Values[0][i].(string), 10, 64)
				if err != nil {
					return err
				}
				reading.Created = n	
			case "origin":
				n, err := strconv.ParseInt(res[0].Series[0].Values[0][i].(string), 10, 64)
				if err != nil {
					return err
				}
				reading.Origin = n	
			case "modified":
				n, err := strconv.ParseInt(res[0].Series[0].Values[0][i].(string), 10, 64)
				if err != nil {
					return err
				}
				reading.Modified = n	
			case "device":
				reading.Device = res[0].Series[0].Values[0][i].(string)
			case "name":
				reading.Name = res[0].Series[0].Values[0][i].(string)
			case "value":
				reading.Value = res[0].Series[0].Values[0][i].(string)
			}
		}
		readings = append(readings, reading)
	}

	ie.Readings = readings

	return nil
}

func queryDB(clnt client.Client, cmd string, db string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: db,
	}
	response, err := clnt.Query(q)
	if err != nil {
		return res, err
	}
	if response.Error() != nil {
		return res, response.Error()
	}
	res = response.Results
	return res, nil
}
