/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
 *******************************************************************************/
package mongo

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

func (m *MongoClient) AddLog(le models.LogEntry) error {

	session := m.session.Copy()
	defer session.Close()

	c := session.DB(m.config.DatabaseName).C(db.LogsCollection)

	if err := c.Insert(le); err != nil {
		return err
	}

	return nil
}

func (m *MongoClient) DeleteLog(criteria db.LogMatcher) (int, error) {

	session := m.session.Copy()
	defer session.Close()

	c := session.DB(m.config.DatabaseName).C(db.LogsCollection)

	base := criteria.CreateQuery()

	info, err := c.RemoveAll(base)

	if err != nil {
		return 0, err
	}

	return info.Removed, nil
}

func (m *MongoClient) FindLog(criteria db.LogMatcher, limit int) ([]models.LogEntry, error) {
	session := m.session.Copy()
	defer session.Close()

	c := session.DB(m.config.DatabaseName).C(db.LogsCollection)

	le := []models.LogEntry{}

	base := criteria.CreateQuery()

	q := c.Find(base)

	if err := q.Limit(limit).All(&le); err != nil {
		return le, err
	}

	return le, nil
}

func (m *MongoClient) ResetLogs() {
	session := m.session.Copy()
	defer session.Close()

	session.DB(m.config.DatabaseName).C(db.LogsCollection).RemoveAll(bson.M{})
	return
}
