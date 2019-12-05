//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package mongo

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/config"
	"strconv"
	"time"

	types "github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/logging/filter"

	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type Logger struct {
	databaseName string
	session      *mgo.Session // Mongo database session
}

/*
	Type    string
	Timeout int
	Host    string
	Port    int
	Name    string

*/

func connectToMongo(
	credentials *types.Credentials,
	host string,
	port int,
	timeout int,
	name string) (*mgo.Session, error) {

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{host + ":" + strconv.Itoa(port)},
		Timeout:  time.Duration(timeout) * time.Millisecond,
		Database: name,
		Username: credentials.Username,
		Password: credentials.Password,
	}

	ms, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return nil, err
	}

	ms.SetSocketTimeout(time.Duration(timeout) * time.Millisecond)
	ms.SetMode(mgo.Monotonic, true)

	return ms, nil
}

func NewLogger(credentials *types.Credentials, configuration *config.ConfigurationStruct) (*Logger, error) {
	databaseName := configuration.Databases["Primary"].Name
	session, err := connectToMongo(
		credentials,
		configuration.Databases["Primary"].Host,
		configuration.Databases["Primary"].Port,
		configuration.Databases["Primary"].Timeout,
		databaseName,
	)
	if err != nil {
		return nil, err
	}
	return &Logger{
		databaseName: databaseName,
		session:      session,
	}, nil
}

func (l *Logger) CloseSession() {
	if l.session != nil {
		l.session.Close()
		l.session = nil
	}
}

func (l *Logger) Add(le models.LogEntry) error {

	session := l.session.Copy()
	defer session.Close()

	c := session.DB(l.databaseName).C(db.LogsCollection)

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

func createQuery(criteria filter.Criteria) bson.M {
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

func (l *Logger) Remove(criteria filter.Criteria) (int, error) {

	session := l.session.Copy()
	defer session.Close()

	c := session.DB(l.databaseName).C(db.LogsCollection)

	base := createQuery(criteria)

	info, err := c.RemoveAll(base)

	if err != nil {
		return 0, err
	}

	return info.Removed, nil
}

func (l *Logger) Find(criteria filter.Criteria) ([]models.LogEntry, error) {
	session := l.session.Copy()
	defer session.Close()

	c := session.DB(l.databaseName).C(db.LogsCollection)

	le := []models.LogEntry{}

	base := createQuery(criteria)

	q := c.Find(base)

	if err := q.Limit(criteria.Limit).All(&le); err != nil {
		return le, err
	}

	return le, nil
}

func (l *Logger) Reset() {
	session := l.session.Copy()
	defer session.Close()

	session.DB(l.databaseName).C(db.LogsCollection).RemoveAll(bson.M{})
	return
}
