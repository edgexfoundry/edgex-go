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
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/globalsign/mgo"
)

var currentMongoClient MongoClient // Singleton used so that mongoEvent can use it to de-reference readings

type MongoClient struct {
	session  *mgo.Session  // Mongo database session
	database *mgo.Database // Mongo database
}

// Return a pointer to the MongoClient
func NewClient(config db.Configuration) (MongoClient, error) {
	m := MongoClient{}

	// Create the dial info for the Mongo session
	connectionString := config.Host + ":" + strconv.Itoa(config.Port)
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{connectionString},
		Timeout:  time.Duration(config.Timeout) * time.Millisecond,
		Database: config.DatabaseName,
		Username: config.Username,
		Password: config.Password,
	}
	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return m, err
	}

	m.session = session
	m.database = session.DB(config.DatabaseName)

	currentMongoClient = m // Set the singleton
	return m, nil
}

func (mc MongoClient) CloseSession() {
	if mc.session != nil {
		mc.session.Close()
		mc.session = nil
	}
}

// Get the current Mongo Client
func getCurrentMongoClient() (MongoClient, error) {
	return currentMongoClient, nil
}

// Get a copy of the session
func (mc MongoClient) getSessionCopy() *mgo.Session {
	return mc.session.Copy()
}

func errorMap(err error) error {
	if err == mgo.ErrNotFound {
		err = db.ErrNotFound
	}
	return err
}

// Delete from the collection based on ID
func (mc MongoClient) deleteById(col string, id string) error {
	s := mc.getSessionCopy()
	defer s.Close()

	query, err := idToBsonM(id)
	if err != nil {
		return err
	}
	return errorMap(s.DB(mc.database.Name).C(col).Remove(query))
}

func (mc MongoClient) updateId(col string, id string, update interface{}) error {
	s := mc.getSessionCopy()
	defer s.Close()

	query, err := idToBsonM(id)
	if err != nil {
		return err
	}
	return errorMap(s.DB(mc.database.Name).C(col).Update(query, update))
}
