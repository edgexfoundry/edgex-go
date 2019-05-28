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
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/gomodule/redigo/redis"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

var currClient *Client // a singleton so Readings can be de-referenced
var once sync.Once

// Client represents a Redis client
type Client struct {
	Pool          *redis.Pool // A thread-safe pool of connections to Redis
	BatchSize     int
	loggingClient logger.LoggingClient
}

type CoreDataClient struct {
	*Client
	logger logger.LoggingClient
}

func NewCoreDataClient(config db.Configuration, logger logger.LoggingClient) (*CoreDataClient, error) {
	var err error
	dc := &CoreDataClient{}
	dc.Client, err = NewClient(config, logger)
	if err != nil {
		return nil, err
	}
	dc.logger = logger
	// Background process for deleting device readings and events.
	// This only needs to be running for core-data since this is the service responsible for handling the deletion
	// of events
	go dc.AsyncDeleteEvents()
	go dc.AsyncDeleteReadings()

	return dc, err
}

// Return a pointer to the Redis client
func NewClient(config db.Configuration, loggingClient logger.LoggingClient) (*Client, error) {
	once.Do(func() {
		connectionString := fmt.Sprintf("%s:%d", config.Host, config.Port)
		opts := []redis.DialOption{
			redis.DialPassword(config.Password),
			redis.DialConnectTimeout(time.Duration(config.Timeout) * time.Millisecond),
		}

		dialFunc := func() (redis.Conn, error) {
			conn, err := redis.Dial(
				"tcp", connectionString, opts...,
			)
			if err != nil {
				return nil, fmt.Errorf("Could not dial Redis: %s", err)
			}
			return conn, nil
		}
		// Default the batch size to 1,000 if not set
		batchSize := 1000
		if config.BatchSize != 0 {
			batchSize = config.BatchSize
		}
		currClient = &Client{
			Pool: &redis.Pool{
				IdleTimeout: 0,
				/* The current implementation processes nested structs using concurrent connections.
				 * With the deepest nesting level being 3, three shall be the number of maximum open
				 * idle connections in the pool, to allow reuse.
				 * TODO: Once we have a concurrent benchmark, this should be revisited.
				 * TODO: Longer term, once the objects are clean of external dependencies, the use
				 * of another serializer should make this moot.
				 */
				MaxIdle: 10,
				Dial:    dialFunc,
			},
			BatchSize:     batchSize,
			loggingClient: loggingClient,
		}
	})
	return currClient, nil
}

// Connect connects to Redis
func (c *Client) Connect() error {
	return nil
}

// CloseSession closes the connections to Redis
func (c *Client) CloseSession() {
	c.Pool.Close()
	close(deleteEventsChannel)
	close(deleteReadingsChannel)
	currClient = nil
	once = sync.Once{}
}

// getConnection gets a connection from the pool
func getConnection() (conn redis.Conn, err error) {
	if currClient == nil {
		return nil, errors.New("No current Redis client: create a new client before getting a connection from it")
	}

	conn = currClient.Pool.Get()
	return conn, nil
}
