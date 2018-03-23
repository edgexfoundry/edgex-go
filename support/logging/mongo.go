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

	conditions := []bson.M{}

	fmt.Println("criteria ", criteria)

	if len(criteria.Labels) > 0 {
		keyCond := []bson.M{}
		for _, label := range criteria.Labels {
			conditions = append(conditions, bson.M{"labels": label})
		}
		conditions = append(conditions, bson.M{"$or": keyCond})
	}

	if len(criteria.Keywords) > 0 {
		keyCond := []bson.M{}
		for _, key := range criteria.Keywords {
			regex := fmt.Sprintf(".*%s.*", key)
			conditions = append(conditions, bson.M{"message": bson.M{"$regex": regex}})
		}
		conditions = append(conditions, bson.M{"$or": keyCond})
	}

	if len(criteria.OriginServices) > 0 {
		keyCond := []bson.M{}
		for _, svc := range criteria.OriginServices {
			regex := fmt.Sprintf(".*%s.*", svc)
			conditions = append(conditions, bson.M{"originservice": bson.M{"$regex": regex}})
		}
		conditions = append(conditions, bson.M{"$or": keyCond})
	}

	if len(criteria.LogLevels) > 0 {
		keyCond := []bson.M{}
		for _, ll := range criteria.LogLevels {
			keyCond = append(keyCond, bson.M{"level": ll})

		}
		conditions = append(conditions, bson.M{"$or": keyCond})
	}

	if criteria.Start != 0 {
		conditions = append(conditions, bson.M{"created": bson.M{"$gt": criteria.Start}})
	}

	if criteria.End != 0 {
		conditions = append(conditions, bson.M{"created": bson.M{"$lt": criteria.End}})
	}

	base := bson.M{"$and": conditions}

	fmt.Println("conditions ", conditions)

	q := c.Find(base)

	if criteria.Limit != 0 {
		q = q.Limit(criteria.Limit)
	}

	if err := q.All(&le); err != nil {
		fmt.Println("Failed to query by id ", err)
		return nil
	}

	fmt.Println("Returned value ", le)

	return le
}

func (ml *mongoLog) reset() {
	fmt.Println("resetting mongos")
}
