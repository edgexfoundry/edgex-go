/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
	"errors"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var currentMongoClient *MongoClient // Singleton used so that mongoEvent can use it to de-reference readings

type MongoClient struct {
	config   db.Configuration
	session  *mgo.Session  // Mongo database session
	database *mgo.Database // Mongo database
}

// Return a pointer to the MongoClient
func NewClient(config db.Configuration) *MongoClient {
	mongoClient := &MongoClient{config: config}
	currentMongoClient = mongoClient // Set the singleton
	return mongoClient
}

func (m *MongoClient) Connect() error {
	// Create the dial info for the Mongo session
	connectionString := m.config.Host + ":" + strconv.Itoa(m.config.Port)
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{connectionString},
		Timeout:  time.Duration(m.config.Timeout) * time.Millisecond,
		Database: m.config.DatabaseName,
		Username: m.config.Username,
		Password: m.config.Password,
	}
	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return err
	}

	m.session = session
	m.database = session.DB(m.config.DatabaseName)
	currentMongoClient = m // Set the singleton

	return nil
}

func (m *MongoClient) CloseSession() {
	if m.session != nil {
		m.session.Close()
		m.session = nil
	}
}

// Get the current Mongo Client
func getCurrentMongoClient() (*MongoClient, error) {
	if currentMongoClient == nil {
		return nil, errors.New("No current mongo client, please create a new client before requesting it")
	}

	return currentMongoClient, nil
}

// Get a copy of the session
func (mc *MongoClient) getSessionCopy() *mgo.Session {
	return mc.session.Copy()
}

func errorMap(err error) error {
	if err == mgo.ErrNotFound {
		err = db.ErrNotFound
	}
	return err
}

// Delete from the collection based on ID
func (mc *MongoClient) deleteById(col string, id string) error {
	// Check if id is a hexstring
	if !bson.IsObjectIdHex(id) {
		return db.ErrInvalidObjectId
	}

	s := mc.getSessionCopy()
	defer s.Close()

	err := s.DB(mc.database.Name).C(col).RemoveId(bson.ObjectIdHex(id))
	return errorMap(err)
}
