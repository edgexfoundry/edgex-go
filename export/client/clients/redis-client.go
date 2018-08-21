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
package clients

import (
	"fmt"
	"time"

	"github.com/edgexfoundry/edgex-go/export"
	"github.com/gomodule/redigo/redis"
	"gopkg.in/mgo.v2/bson"
)

type connectionType int

const (
	connTypeTCP connectionType = 1 << iota
	connTypeUDS
	connTypeEredis
)

const (
	protoTCP    = "tcp"
	protoUDS    = "unix"
	protoEredis = "eredis"
)

var currClients = make([]*Client, 3) // A singleton per connection type (for benchmarking)
var currClient *Client               // a singleton so Readings can be de-referenced

// Client represents a client
type Client struct {
	Pool     *redis.Pool // Connections to Redis
	conntype connectionType
}

// Return a pointer to the RedisClient
func newRedisClient(config DBConfiguration) (*Client, error) {
	// Identify the connection's type
	var conntype connectionType
	if config.Host == "" {
		conntype = connTypeEredis
	} else if string(config.Host[0]) == "/" {
		conntype = connTypeUDS
	} else {
		conntype = connTypeTCP
	}

	if currClients[conntype] == nil {
		connectionString := fmt.Sprintf("%s:%d", config.Host, config.Port)

		var proto, addr string
		c := Client{
			conntype: conntype,
		}
		switch c.conntype {
		case connTypeEredis:
			proto = protoEredis
		case connTypeUDS:
			proto = protoUDS
			addr = config.Host
		case connTypeTCP:
			proto = protoTCP
			addr = connectionString
		default:
			return nil, ErrUnsupportedDatabase
		}

		opts := []redis.DialOption{
			redis.DialPassword(config.Password),
			redis.DialConnectTimeout(time.Duration(config.Timeout) * time.Millisecond),
		}

		dialFunc := func() (redis.Conn, error) {
			conn, err := redis.Dial(
				proto, addr, opts...,
			)
			if err != nil {
				return nil, err
			}
			return conn, nil
		}

		c.Pool = &redis.Pool{
			MaxIdle:     1,
			IdleTimeout: 0,
			Dial:        dialFunc,
		}

		currClients[conntype] = &c
		currClient = &c
	} else {
		currClient = currClients[conntype]
	}

	return currClient, nil
}

// Connect connects to Redis
func (c *Client) Connect() error {
	return nil
}

// CloseSession closes the connections to Redis
func (c *Client) CloseSession() {
	c.Pool.Close()
	currClients[c.conntype] = nil
	currClient = nil
}

// ********************** REGISTRATION FUNCTIONS *****************************
// Return all the registrations
// UnexpectedError - failed to retrieve registrations from the database
func (c *Client) Registrations() (r []export.Registration, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, EXPORT_COLLECTION, 0, -1)
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

	err = getObjectByHash(conn, EXPORT_COLLECTION+":name", name, unmarshalObject, &r)
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

	id, err := redis.String(conn.Do("HGET", EXPORT_COLLECTION+":name", name))
	if err == redis.ErrNil {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	return deleteRegistration(conn, id)
}

func (c *Client) ScrubAllExports() (err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	return unlinkCollection(conn, EXPORT_COLLECTION)
}

func addRegistration(conn redis.Conn, r *export.Registration) (id bson.ObjectId, err error) {
	if !r.ID.Valid() {
		r.ID = bson.NewObjectId()
	}
	id = r.ID
	rid := r.ID.Hex()

	ts := time.Now().UnixNano() / int64(time.Millisecond) //MakeTimestamp()
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
	conn.Send("ZADD", EXPORT_COLLECTION, 0, rid)
	conn.Send("HSET", EXPORT_COLLECTION+":name", r.Name, rid)
	_, err = conn.Do("EXEC")
	return id, err
}

func deleteRegistration(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	r := export.Registration{}
	err = unmarshalObject(object, &r)

	conn.Send("MULTI")
	conn.Send("DEL", id)
	conn.Send("ZREM", EXPORT_COLLECTION, id)
	conn.Send("HDEL", EXPORT_COLLECTION+":name", r.Name)
	_, err = conn.Do("EXEC")
	return err
}

// ------------------------------------ "imports" ------------------------------
const (
	scriptUnlinkZsetMembers = "unlinkZsetMembers"
	scriptUnlinkCollection  = "unlinkCollection"
)

var scripts = map[string]redis.Script{
	scriptUnlinkZsetMembers: *redis.NewScript(1, `
		local magic = 4096
		local ids = redis.call('ZRANGE', KEYS[1], 0, -1)
		if #ids > 0 then
			for i = 1, #ids, magic do
				redis.call('UNLINK', unpack(ids, i, i+magic < #ids and i+magic or #ids))
			end
		end
		`),
	scriptUnlinkCollection: *redis.NewScript(0, `
		local magic = 4096
		redis.replicate_commands()
		local c = 0
		repeat
			local s = redis.call('SCAN', c, 'MATCH', ARGV[1] .. '*')
			c = tonumber(s[1])
			if #s[2] > 0 then
				redis.call('UNLINK', unpack(s[2]))
			end
		until c == 0
		`),
}

type marshalFunc func(in interface{}) (out []byte, err error)
type unmarshalFunc func(in []byte, out interface{}) (err error)

func marshalObject(in interface{}) (out []byte, err error) {
	return bson.Marshal(in)
}

func unmarshalObject(in []byte, out interface{}) (err error) {
	return bson.Unmarshal(in, out)
}

func getObjectsByRange(conn redis.Conn, key string, start, end int) (objects [][]byte, err error) {
	ids, err := redis.Values(conn.Do("ZRANGE", key, start, end))
	if err != nil && err != redis.ErrNil {
		return nil, err
	}

	if len(ids) > 0 {
		objects, err = redis.ByteSlices(conn.Do("MGET", ids...))
		if err != nil {
			return nil, err
		}
	}
	return objects, nil
}

func getObjectById(conn redis.Conn, id string, unmarshal unmarshalFunc, out interface{}) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	return unmarshal(object, out)
}

func getObjectByHash(conn redis.Conn, hash string, field string, unmarshal unmarshalFunc, out interface{}) error {
	id, err := redis.String(conn.Do("HGET", hash, field))
	if err == redis.ErrNil {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	object, err := redis.Bytes(conn.Do("GET", id))
	if err != nil {
		return err
	}

	return unmarshal(object, out)
}

func unlinkCollection(conn redis.Conn, col string) error {
	conn.Send("MULTI")
	s := scripts[scriptUnlinkZsetMembers]
	s.Send(conn, col)
	s = scripts[scriptUnlinkCollection]
	s.Send(conn, col)
	_, err := conn.Do("EXEC")
	return err
}
