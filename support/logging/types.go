//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"os"
	"strconv"

	support_domain "github.com/edgexfoundry/edgex-go/support/domain"
)

const (
	applicationName = "support-logging"
	defaultPort     = 48061
	//defaultPersistence = PersistenceFile
	defaultPersistence = PersistenceMongo
	defaultLogFilename = "support-logging.log"

	defaultMongoDB             = "logging"
	defaultMongoCollection     = "logEntry"
	defaultMongoURL            = "127.0.0.1"
	defaultMongoPort           = 27017
	defaultMongoConnectTimeout = 5000
	defaultSocketTimeout       = 5000
	defaultMongoUsername       = "logging"
	defaultMongoPassword       = "password"

	envMongoURL  = "SUPPORT_LOGGING_MONGO_URL"
	envMongoPort = "SUPPORT_LOGGING_MONGO_PORT"

	PersistenceMongo = "mongodb"
	PersistenceFile  = "file"
)

type Config struct {
	Port        int
	Persistence string

	// Used by PersistenceFile
	LogFilename string

	// Used by mongo
	MongoURL            string
	MongoUser           string
	MongoPass           string
	MongoDatabase       string
	MongoCollection     string
	MongoPort           int
	MongoConnectTimeout int
	MongoSocketTimeout  int
}

type persistence interface {
	add(logEntry support_domain.LogEntry)
	remove(criteria matchCriteria) int
	find(criteria matchCriteria) []support_domain.LogEntry

	// Needed for the tests. Reset the instance (closing files, sessions...)
	// and clear the logs.
	reset()
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func GetDefaultConfig() Config {
	cfg := Config{
		Port:        defaultPort,
		Persistence: defaultPersistence,
		LogFilename: defaultLogFilename,

		MongoURL: env(envMongoURL, defaultMongoURL),

		MongoUser:           defaultMongoUsername,
		MongoPass:           defaultMongoPassword,
		MongoDatabase:       defaultMongoDB,
		MongoCollection:     defaultMongoCollection,
		MongoConnectTimeout: defaultMongoConnectTimeout,
		MongoSocketTimeout:  defaultSocketTimeout,
	}

	MongoPortStr := env(envMongoPort, strconv.Itoa(defaultMongoPort))
	MongoPort, err := strconv.Atoi(MongoPortStr)
	if err == nil {
		cfg.MongoPort = MongoPort
	} else {
		cfg.MongoPort = defaultMongoPort
	}

	return cfg
}
