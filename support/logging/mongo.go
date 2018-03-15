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

	support_domain "github.com/edgexfoundry/edgex-go/support/domain"

	mgo "gopkg.in/mgo.v2"
	bson "gopkg.in/mgo.v2/bson"
)

type mongoLog struct {
	session *mgo.Session // Mongo database session
	config  *Config
}

func connectToMongo(cfg *Config) (*mgo.Session, error) {
	fmt.Println("connecting mongos")

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

func (ml *mongoLog) add(le support_domain.LogEntry) {

	session := ml.session.Copy()
	defer session.Close()

	c := session.DB(ml.config.MongoDatabase).C(ml.config.MongoCollection)

	if err := c.Insert(le); err != nil {
		fmt.Println("Failed to add log", err)
		return
	}
	fmt.Println("adding mongos")
}

func (ml *mongoLog) remove(criteria matchCriteria) int {

	fmt.Println("removing mongos")
	return 0
}

func (ml *mongoLog) find(criteria matchCriteria) []support_domain.LogEntry {
	fmt.Println("finding mongos")
	session := ml.session.Copy()
	defer session.Close()

	c := session.DB(ml.config.MongoDatabase).C(ml.config.MongoCollection)

	le := []support_domain.LogEntry{}

	//type M map[string]interface{}

	//query := bson.M{"created": bson.M{"$gt": criteria.Start}}
	fmt.Print(criteria)
	if err := c.Find(bson.M{"created": bson.M{"$gt": 1}}).All(&le); err != nil {
		fmt.Println("Failed to query by id %v", err)
		return nil
	}

	fmt.Println("Returned value %v", le)

	return le
}

func (ml *mongoLog) reset() {
	fmt.Println("resetting mongos")
}
