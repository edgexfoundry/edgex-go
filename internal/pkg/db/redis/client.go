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
	"os"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"

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
}

// Return a pointer to the Redis client
func NewClient(config db.Configuration, lc logger.LoggingClient) (*Client, error) {
	var retErr error
	once.Do(func() {
		connectionString := fmt.Sprintf("%s:%d", config.Host, config.Port)
		connectTimeout, err := time.ParseDuration(config.Timeout)
		if err != nil {
			retErr = fmt.Errorf("configured database timeout failed to parse: %v", err)
			return
		}
		opts := []redis.DialOption{
			redis.DialConnectTimeout(connectTimeout),
		}
		if os.Getenv("EDGEX_SECURITY_SECRET_STORE") != "false" {
			opts = append(opts, redis.DialPassword(config.Password))
		}

		dialFunc := func() (redis.Conn, error) {
			conn, err := redis.Dial(
				"tcp", connectionString, opts...,
			)
			if err == nil {
				_, err = conn.Do("PING")
				if err == nil {
					return conn, nil
				}
			}

			return nil, fmt.Errorf("could not dial Redis: %s", err)
		}
		// Default the batch size to 1,000 if not set
		batchSize := 1000
		if config.BatchSize != 0 {
			batchSize = config.BatchSize
		}
		currClient = &Client{
			Pool: &redis.Pool{
				IdleTimeout: connectTimeout,
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
			loggingClient: lc,
		}
	})

	// Test connectivity now so don't have failures later when doing lazy connect.
	if _, err := currClient.Pool.Dial(); err != nil {
		return nil, err
	}

	return currClient, retErr
}

// Connect connects to Redis
func (c *Client) Connect() error {
	return nil
}

// CloseSession closes the connections to Redis
func (c *Client) CloseSession() {
	_ = c.Pool.Close()
	currClient = nil
	once = sync.Once{}
}
