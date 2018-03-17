//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

const (
	applicationName    = "support-scheduler"
	defaultPort        = 48061
	defaultPersistence = PersistenceFile

	PersistenceMongo = "mongodb"
	PersistenceFile  = "file"
)

type Config struct {
	Port        int
	Persistence string
}

func GetDefaultConfig() Config {
	return Config{
		Port:        defaultPort,
		Persistence: defaultPersistence,
	}
}
