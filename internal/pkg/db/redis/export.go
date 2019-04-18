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
	"github.com/google/uuid"
)

// ********************** REGISTRATION FUNCTIONS *****************************
// Return all the registrations
// UnexpectedError - failed to retrieve registrations from the database
func (c *Client) Registrations() (r []contract.Registration, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.ExportCollection, 0, -1)
	if err != nil {
		return nil, err
	}

	r = make([]contract.Registration, len(objects))
	for i, object := range objects {
		err = json.Unmarshal(object, &r[i])
		if err != nil {
			return nil, err
		}
	}

	return r, err
}

// Add a new registration
// UnexpectedError - failed to add to database
func (c *Client) AddRegistration(reg contract.Registration) (id string, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	if reg.ID != "" {
		_, err = uuid.Parse(reg.ID)
		if err != nil {
			return "", db.ErrInvalidObjectId
		}
	}

	return addRegistration(conn, reg)
}

// Update a registration
// UnexpectedError - problem updating in database
// NotFound - no registration with the ID was found
func (c *Client) UpdateRegistration(reg contract.Registration) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteRegistration(conn, reg.ID)
	if err != nil {
		return err
	}

	_, err = addRegistration(conn, reg)
	return err
}

// Get a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (c *Client) RegistrationById(id string) (r contract.Registration, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	err = getObjectById(conn, id, unmarshalObject, &r)
	return r, err
}

// Get a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the name was found
func (c *Client) RegistrationByName(name string) (r contract.Registration, err error) {
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

func addRegistration(conn redis.Conn, r contract.Registration) (id string, err error) {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}

	ts := db.MakeTimestamp()
	if r.Created == 0 {
		r.Created = ts
	}
	r.Modified = ts

	m, err := marshalObject(r)
	if err != nil {
		return r.ID, err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("SET", r.ID, m)
	_ = conn.Send("ZADD", db.ExportCollection, 0, r.ID)
	_ = conn.Send("HSET", db.ExportCollection+":name", r.Name, r.ID)
	_, err = conn.Do("EXEC")
	return r.ID, err
}

func deleteRegistration(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	r := contract.Registration{}
	_ = unmarshalObject(object, &r)

	_ = conn.Send("MULTI")
	_ = conn.Send("DEL", id)
	_ = conn.Send("ZREM", db.ExportCollection, id)
	_ = conn.Send("HDEL", db.ExportCollection+":name", r.Name)
	_, err = conn.Do("EXEC")
	return err
}
