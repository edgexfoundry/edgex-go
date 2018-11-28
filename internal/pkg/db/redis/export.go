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
	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/gomodule/redigo/redis"
	"gopkg.in/mgo.v2/bson"
)

// ********************** REGISTRATION FUNCTIONS *****************************
// Return all the registrations
// UnexpectedError - failed to retrieve registrations from the database
func (c *Client) Registrations() (r []export.Registration, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.ExportCollection, 0, -1)
	if err != nil {
		return nil, err
	}

	r = make([]export.Registration, len(objects))
	for i, object := range objects {
		err = bson.Unmarshal(object, &r[i])
		if err != nil {
			return nil, err
		}
	}

	return r, err
}

// Add a new registration
// UnexpectedError - failed to add to database
func (c *Client) AddRegistration(reg *export.Registration) (id bson.ObjectId, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	return addRegistration(conn, reg)
}

// Update a registration
// UnexpectedError - problem updating in database
// NotFound - no registration with the ID was found
func (c *Client) UpdateRegistration(reg export.Registration) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteRegistration(conn, reg.ID.Hex())
	if err != nil {
		return err
	}

	_, err = addRegistration(conn, &reg)
	return err
}

// Get a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (c *Client) RegistrationById(id string) (r export.Registration, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	err = getObjectById(conn, id, unmarshalObject, &r)
	return r, err
}

// Get a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the name was found
func (c *Client) RegistrationByName(name string) (r export.Registration, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	err = getObjectByHash(conn, db.ExportCollection+":name", name, unmarshalObject, &r)
	return r, err
}

// Delete a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (c *Client) DeleteRegistrationById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteRegistration(conn, id)
}

// Delete a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the name was found
func (c *Client) DeleteRegistrationByName(name string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	id, err := redis.String(conn.Do("HGET", db.ExportCollection+":name", name))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	return deleteRegistration(conn, id)
}

//  ScrubAllRegistrations deletes all export related data
func (c *Client) ScrubAllRegistrations() (err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	return unlinkCollection(conn, db.ExportCollection)
}

func addRegistration(conn redis.Conn, r *export.Registration) (id bson.ObjectId, err error) {
	if !r.ID.Valid() {
		r.ID = bson.NewObjectId()
	}
	id = r.ID
	rid := r.ID.Hex()

	ts := db.MakeTimestamp()
	if r.Created == 0 {
		r.Created = ts
	}
	r.Modified = ts

	m, err := marshalObject(r)
	if err != nil {
		return id, err
	}

	conn.Send("MULTI")
	conn.Send("SET", rid, m)
	conn.Send("ZADD", db.ExportCollection, 0, rid)
	conn.Send("HSET", db.ExportCollection+":name", r.Name, rid)
	_, err = conn.Do("EXEC")
	return id, err
}

func deleteRegistration(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	r := export.Registration{}
	err = unmarshalObject(object, &r)

	conn.Send("MULTI")
	conn.Send("DEL", id)
	conn.Send("ZREM", db.ExportCollection, id)
	conn.Send("HDEL", db.ExportCollection+":name", r.Name)
	_, err = conn.Do("EXEC")
	return err
}
