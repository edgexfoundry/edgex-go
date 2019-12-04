//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/interfaces"

	mgo "github.com/globalsign/mgo"
	bson "github.com/globalsign/mgo/bson"
)

type mongoLog struct {
	session *mgo.Session // Mongo database session
}

func connectToMongo(credentials config.Credentials) (*mgo.Session, error) {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{Configuration.Databases["Primary"].Host + ":" + strconv.Itoa(Configuration.Databases["Primary"].Port)},
		Timeout:  time.Duration(Configuration.Databases["Primary"].Timeout) * time.Millisecond,
		Database: Configuration.Databases["Primary"].Name,
		Username: credentials.Username,
		Password: credentials.Password,
	}

	ms, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return nil, err
	}

	ms.SetSocketTimeout(time.Duration(Configuration.Databases["Primary"].Timeout) * time.Millisecond)
	ms.SetMode(mgo.Monotonic, true)

	return ms, nil
}

func (ml *mongoLog) CloseSession() {
	if ml.session != nil {
		ml.session.Close()
		ml.session = nil
	}
}

func (ml *mongoLog) Add(le models.LogEntry) error {

	session := ml.session.Copy()
	defer session.Close()

	c := session.DB(Configuration.Databases["Primary"].Name).C(db.LogsCollection)

	if err := c.Insert(le); err != nil {
		return err
	}

	return nil
}

func createConditions(conditions []bson.M, field string, elements []string) []bson.M {
	keyCond := []bson.M{}
	for _, value := range elements {
		keyCond = append(keyCond, bson.M{field: value})
	}

	return append(conditions, bson.M{"$or": keyCond})
}

func createQuery(criteria MatchCriteria) bson.M {
	conditions := []bson.M{{}}

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

func (ml *mongoLog) Remove(criteria interfaces.Criteria) (int, error) {
	session := ml.session.Copy()
	defer session.Close()

	c := session.DB(Configuration.Databases["Primary"].Name).C(db.LogsCollection)

	matchCriteria, ok := criteria.(MatchCriteria)
	if !ok {
		return 0, errors.New("unknown criteria type")
	}
	base := createQuery(matchCriteria)

	info, err := c.RemoveAll(base)

	if err != nil {
		return 0, err
	}

	return info.Removed, nil
}

func (ml *mongoLog) Find(criteria interfaces.Criteria) ([]models.LogEntry, error) {
	session := ml.session.Copy()
	defer session.Close()

	c := session.DB(Configuration.Databases["Primary"].Name).C(db.LogsCollection)

	matchCriteria, ok := criteria.(MatchCriteria)
	if !ok {
		return nil, errors.New("unknown criteria type")
	}

	base := createQuery(matchCriteria)

	q := c.Find(base)

	le := []models.LogEntry{}
	if err := q.Limit(matchCriteria.Limit).All(&le); err != nil {
		return le, err
	}

	return le, nil
}
