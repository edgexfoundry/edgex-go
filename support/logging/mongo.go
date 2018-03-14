//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"fmt"
	"strconv"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/edgexfoundry/edgex-go/support/domain"
)

type DataStore struct {
	s *mgo.Session
}

// Repository - get Mongo session
type Repository struct {
	Session *mgo.Session
}

type mongoLog struct {
	session *mgo.Session // Mongo database session
}

func connectToMongo(cfg *Config) (*mgo.Session, error) {
	fmt.Println("connectin mongos")

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{cfg.MongoURL + ":" + strconv.Itoa(cfg.MongoPort)},
		Timeout:  time.Duration(cfg.MongoConnectTimeout) * time.Millisecond,
		Database: cfg.MongoDatabase,
		Username: cfg.MongoUser,
		Password: cfg.MongoPass,
	}

	ms, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return nil, err
	}

	ms.SetSocketTimeout(time.Duration(cfg.MongoSocketTimeout) * time.Millisecond)
	ms.SetMode(mgo.Monotonic, true)

	return ms, nil
}

// NewRepository - create new Mongo repository
func NewRepository(ms *mgo.Session) *Repository {
	return &Repository{Session: ms}
}

func (ml *mongoLog) add(le support_domain.LogEntry) {
	fmt.Println("adding mongos")
}

func (ml *mongoLog) remove(criteria matchCriteria) int {
	fmt.Println("removing mongos")
	return 0
}

func (ml *mongoLog) find(criteria matchCriteria) []support_domain.LogEntry {
	fmt.Println("finding mongos")
	return nil
}

func (ml *mongoLog) reset() {
	fmt.Println("resetting mongos")
}
