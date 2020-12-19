/*******************************************************************************
* Copyright 2020 Redis Labs
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
* @author: Andre Srinivasan
*******************************************************************************/

package redis

import (
	"context"
	"fmt"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/redis"
	"github.com/edgexfoundry/edgex-go/internal/security/redis/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	redigo "github.com/gomodule/redigo/redis"
)

type Handler struct {
	secured     bool
	credentials bootstrapConfig.Credentials
	redisConn   redigo.Conn
}

func NewHandler() *Handler {
	return &Handler{}
}

func (handler *Handler) getCredentials(ctx context.Context, _ *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)

	var credentials = bootstrapConfig.Credentials{
		Username: "unset",
		Password: "unset",
	}

	for startupTimer.HasNotElapsed() {
		secrets, err := secretProvider.GetSecrets(config.Databases["Primary"].Type)
		if err == nil {
			credentials.Username = secrets[secret.UsernameKey]
			credentials.Password = secrets[secret.PasswordKey]
			break
		}

		lc.Warn(fmt.Sprintf("Could not retrieve database credentials (startup timer has not expired): %s", err.Error()))
		startupTimer.SleepForInterval()
	}

	if credentials.Password == "unset" {
		lc.Error("Failed to retrieve database credentials before startup timer expired")
		return false
	}

	handler.credentials = credentials
	return true
}

func (handler *Handler) connect(ctx context.Context, _ *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

	var redisConn redigo.Conn

	for startupTimer.HasNotElapsed() {
		lc.Debug("Attempting unauthenticated connection")
		redisClient, err := redis.NewClient(db.Configuration{
			Host: config.Databases["Primary"].Host,
			Port: config.Databases["Primary"].Port,
		}, lc)
		if err != nil {
			lc.Warn(fmt.Sprintf("Could not create database client (startup timer has not expired): %s", err.Error()))
			startupTimer.SleepForInterval()
			continue
		}

		redisConn = redisClient.Pool.Get()
		if err = testConnection(redisConn); err == nil {
			break
		}
		lc.Debug("Unauthenticated connection failed. Attempting authenticated connection.")

		if err = authenticate(redisConn, handler.credentials); err == nil {
			if err = testConnection(redisConn); err == nil {
				handler.secured = true
				break
			}
		}

		redisConn = nil
		lc.Debug("Authenticated connected failed.")

		lc.Warn(fmt.Sprintf("Could not create database client (startup timer has not expired): %s", err.Error()))
		startupTimer.SleepForInterval()
	}

	if redisConn == nil {
		lc.Error("Failed to create database client before startup timer expired")
		return false
	}

	lc.Info("Connected to database.")
	handler.redisConn = redisConn

	return true
}

func (handler *Handler) maybeSetCredentials(ctx context.Context, _ *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	if !handler.secured {
		_, err := handler.redisConn.Do("CONFIG", "SET", "REQUIREPASS", handler.credentials.Password)
		if err != nil {
			lc.Error(fmt.Sprintf("Could not set Redis password: %s", err.Error()))
			return false
		}

		handler.secured = true
		lc.Info("Database credentials have been set.")
	}

	if err := testConnection(handler.redisConn); err != nil {
		lc.Error(fmt.Sprintf("Connection verification failed: %s", err.Error()))
		return false
	}

	lc.Info("Connection verified.")
	return true
}

func testConnection(redisConn redigo.Conn) error {
	_, err := redisConn.Do("INFO", "SERVER")
	return err
}

func authenticate(redisConn redigo.Conn, credentials bootstrapConfig.Credentials) error {
	_, err := redisConn.Do("AUTH", credentials.Password)
	return err
}
