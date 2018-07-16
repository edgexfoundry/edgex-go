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
}

func connectToMongo() (*mgo.Session, error) {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{configuration.MongoURL + ":" + strconv.Itoa(configuration.MongoPort)},
		Timeout:  time.Duration(configuration.MongoConnectTimeout) * time.Millisecond,
		Database: configuration.MongoDB,
		Username: configuration.MongoUsername,
		Password: configuration.MongoPassword,
	}

	ms, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return nil, err
	}

	ms.SetSocketTimeout(time.Duration(configuration.SocketTimeout) * time.Millisecond)
	ms.SetMode(mgo.Monotonic, true)

	return ms, nil
}

func (ml *mongoLog) add(le support_domain.LogEntry) {

	session := ml.session.Copy()
	defer session.Close()

	c := session.DB(configuration.MongoDB).C(configuration.MongoCollection)

	if err := c.Insert(le); err != nil {
		return
	}
}

func createConditions(conditions []bson.M, field string, elements []string) []bson.M {
	keyCond := []bson.M{}
	for _, value := range elements {
		keyCond = append(keyCond, bson.M{field: value})
	}

	return append(conditions, bson.M{"$or": keyCond})
}

func createQuery(criteria matchCriteria) bson.M {
	conditions := []bson.M{{}}

	if len(criteria.Labels) > 0 {
		conditions = createConditions(conditions, "labels", criteria.Labels)
	}

	if len(criteria.Keywords) > 0 {
		keyCond := []bson.M{}
		for _, key := range criteria.Keywords {
			regex := fmt.Sprintf(".*%s.*", key)
			keyCond = append(keyCond, bson.M{"message": bson.M{"$regex": regex}})
		}
		conditions = append(conditions, bson.M{"$or": keyCond})
	}

	if len(criteria.OriginServices) > 0 {
		conditions = createConditions(conditions, "originService", criteria.OriginServices)
	}

	if len(criteria.LogLevels) > 0 {
		conditions = createConditions(conditions, "logLevel", criteria.LogLevels)
	}

	if criteria.Start != 0 {
		conditions = append(conditions, bson.M{"created": bson.M{"$gt": criteria.Start}})
	}

	if criteria.End != 0 {
		conditions = append(conditions, bson.M{"created": bson.M{"$lt": criteria.End}})
	}

	return bson.M{"$and": conditions}

}

func (ml *mongoLog) remove(criteria matchCriteria) int {

	session := ml.session.Copy()
	defer session.Close()

	c := session.DB(configuration.MongoDB).C(configuration.MongoCollection)

	base := createQuery(criteria)

	info, err := c.RemoveAll(base)

	if err != nil {
		return 0
	}

	return info.Removed
}

func (ml *mongoLog) find(criteria matchCriteria) []support_domain.LogEntry {
	session := ml.session.Copy()
	defer session.Close()

	c := session.DB(configuration.MongoDB).C(configuration.MongoCollection)

	le := []support_domain.LogEntry{}

	base := createQuery(criteria)

	q := c.Find(base)

	if err := q.Limit(criteria.Limit).All(&le); err != nil {
		return nil
	}

	return le
}

func (ml *mongoLog) reset() {
	session := ml.session.Copy()
	defer session.Close()

	session.DB(configuration.MongoDB).C(configuration.MongoCollection).RemoveAll(bson.M{})
	return
}
