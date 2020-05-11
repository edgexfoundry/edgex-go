// +build redisIntegration

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

// This test will only be executed if the tag redisRunning is added when running
// the tests with a command like:
// LD_LIBRARY_PATH=$GOROOT/src/github.com/redislab/eredis/redis/src go test -tags redisRunning

// To test Redis, specify the a `Host` value as follows:
// * TCP connection: use the IP address or host name
// * Unix domain socket: use the path to the socket file (e.g. /tmp/redis.sock)
// * Embedded: leave empty

package redis

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/test"
)

const (
	RedisURLEnvName = "REDIS_SERVER_TEST"
	DefaultRedisURL = "redis://localhost:6379"
)

func TestRedisDB(t *testing.T) {
	config, err := getDBConfiguration()
	require.NoError(t, err)
	rc, err := NewClient(config, logger.MockLogger{})
	if err != nil {
		t.Fatalf("Could not connect with Redis: %v", err)
	}
	test.TestDataDB(t, rc)
	rc.CloseSession()

	rc, err = NewClient(config, logger.MockLogger{})
	if err != nil {
		t.Fatalf("Could not connect with Redis: %v", err)
	}
	test.TestMetadataDB(t, rc)
	rc.CloseSession()

	rc, err = NewClient(config, logger.MockLogger{})
	if err != nil {
		t.Fatalf("Could not connect with Redis: %v", err)
	}

	rc, err = NewClient(config, logger.MockLogger{})
	if err != nil {
		t.Fatalf("Could not connect with Redis: %v", err)
	}
	test.TestNotificationsDB(t, rc)
	rc.CloseSession()

}

func BenchmarkRedisDB(b *testing.B) {
	config, err := getDBConfiguration()
	require.NoError(b, err)
	rc, err := NewClient(config, logger.MockLogger{})
	if err != nil {
		b.Fatalf("Could not connect with Redis: %v", err)
	}

	test.BenchmarkDB(b, rc)
}

func getDBConfiguration() (db.Configuration, error) {
	redisURLString := os.Getenv(RedisURLEnvName)
	if redisURLString == "" {
		redisURLString = DefaultRedisURL
	}

	redisURL, err := url.Parse(redisURLString)
	if err != nil {
		return db.Configuration{}, errors.New(fmt.Sprintf("unable to parse provided Redis URL '%s'", redisURLString))
	}

	portInt, err := strconv.Atoi(redisURL.Port())
	if err != nil {
		return db.Configuration{}, errors.New(fmt.Sprintf("unable to parse provided Redis Port '%s'", redisURLString))
	}

	return db.Configuration{
		Host: redisURL.Hostname(),
		Port: portInt,
	}, nil
}
