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
	"errors"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/gomodule/redigo/redis"
)

type connectionType int

const (
	conntypeTCP connectionType = 1 << iota
	conntypeUDS
	conntypeEredis
)

var currClients = make([]*Client, 3) // A singleton per connection type (for benchmarking)
var currClient *Client               // a singleton so Readings can be de-referenced

// Client represents a client
type Client struct {
	Pool     *redis.Pool // Connections to Redis
	conntype connectionType
}

// Return a pointer to the RedisClient
func NewClient(config db.Configuration) (*Client, error) {
	// Identify the connection's type
	var conntype connectionType
	if config.Host == "" {
		conntype = conntypeEredis
	} else if string(config.Host[0]) == "/" {
		conntype = conntypeUDS
	} else {
		conntype = conntypeTCP
	}

	if currClients[conntype] == nil {
		connectionString := config.Host + ":" + strconv.Itoa(config.Port)

		var proto, addr string
		c := Client{
			conntype: conntype,
		}
		switch c.conntype {
		case conntypeEredis:
			proto = "eredis"
			addr = ""
		case conntypeUDS:
			proto = "unix"
			addr = config.Host
		case conntypeTCP:
			proto = "tcp"
			addr = connectionString
		default:
			return nil, db.ErrUnsupportedDatabase
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
			MaxIdle:     10,
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

// GetConnection returns a Redis connection from the current client's pool
func GetConnection() (conn redis.Conn, err error) {
	if currClient == nil {
		return nil, errors.New("No current Redis client, please create a new client before requesting it")
	}

	conn = currClient.Pool.Get()
	return conn, nil
}
