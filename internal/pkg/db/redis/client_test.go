// +build redisRunning

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
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/test"
)

func TestRedisDB(t *testing.T) {

	t.Log("This test needs to have a running Redis on localhost")

	config := db.Configuration{
		Host: "0.0.0.0",
		Port: 6379,
	}

	rc, err := NewClient(config)
	if err != nil {
		t.Fatalf("Could not connect with Redis: %v", err)
	}
	test.TestDataDB(t, rc)

	rc, err = NewClient(config)
	if err != nil {
		t.Fatalf("Could not connect with Redis: %v", err)
	}

	rc.ScrubAllMetadata()
	test.TestMetadataDB(t, rc)
}

func BenchmarkRedisDB_TCP(b *testing.B) {

	b.Log("This benchmark needs to have a running Redis on localhost")

	config := db.Configuration{
		Host: "0.0.0.0",
		Port: 6379,
	}

	rc, err := NewClient(config)
	if err != nil {
		b.Fatalf("Could not connect with Redis: %v", err)
	}

	test.BenchmarkDB(b, rc)
}

func BenchmarkRedisDB_UDS(b *testing.B) {

	b.Log("This benchmark needs to have a running Redis listening /tmp/redis.sock")

	config := db.Configuration{
		Host: "/tmp/redis.sock",
	}

	rc, err := NewClient(config)
	if err != nil {
		b.Fatalf("Could not connect with Redis: %v", err)
	}

	test.BenchmarkDB(b, rc)
}

func BenchmarkRedisDB_Embedded(b *testing.B) {

	b.Log("This benchmark needs to have liberedis.so at LD_LIBRARY_PATH=$GOROOT/src/github.com/redislabs/eredis/redis/src")

	config := db.Configuration{
		Host: "",
	}

	rc, err := NewClient(config)
	if err != nil {
		b.Fatalf("Could not connect with Redis: %v", err)
	}

	test.BenchmarkDB(b, rc)
}
