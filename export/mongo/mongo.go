//
// Copyright (c) 2017 Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package mongo

import (
	"gopkg.in/mgo.v2"
)

// DBName - DB name
// CollectionName - Collection name
const (
	DBName         string = "coredata"
	CollectionName string = "exportConfiguration"
)

// Repository - get Mongo session
type Repository struct {
	Session *mgo.Session
}

// NewRepository - create new Mongo repository
func NewRepository(ms *mgo.Session) *Repository {
	return &Repository{Session: ms}
}
