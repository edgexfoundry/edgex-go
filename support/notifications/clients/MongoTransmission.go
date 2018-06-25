/*******************************************************************************
* Copyright 2018 Dell Technologies Inc.
*
* Licensed under the Apache License, Version 2.0 (the "License"); you may not us
* in compliance with the License. You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software distribute
* is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KI
* or implied. See the License for the specific language governing permissions an
* the License.
*
*******************************************************************************/
package clients

import (
	"github.com/edgexfoundry/edgex-go/support/notifications/models"
	"gopkg.in/mgo.v2/bson"
)

// Struct that wraps an event to handle DBRefs
type MongoTransmission struct {
	models.Transmission
}

// Custom marshaling into mongo
func (mt MongoTransmission) GetBSON() (interface{}, error) {

	return struct {
		ID           bson.ObjectId               `bson:"_id,omitempty"`
		Modified     int64                       `bson:"modified"`
		Created      int64                       `bson:"created"`
		Notification models.Notification         `bson:"notification,omitempty"`
		Receiver     string                      `bson:"receiver"`
		Channel      models.Channel              `bson:"channel"`
		Status       models.TransmissionStatus   `bson:"status"`
		ResendCount  int                         `bson:"resendcount"`
		Records      []models.TransmissionRecord `bson:"records,omitempty"`
	}{
		ID:           mt.ID,
		Created:      mt.Created,
		Modified:     mt.Modified,
		Notification: mt.Notification,
		Receiver:     mt.Receiver,
		Channel:      mt.Channel,
		Status:       mt.Status,
		ResendCount:  mt.ResendCount,
		Records:      mt.Records,
	}, nil
}

// Custom unmarshaling out of mongo
func (mt *MongoTransmission) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		ID           bson.ObjectId               `bson:"_id,omitempty"`
		Modified     int64                       `bson:"modified"`
		Created      int64                       `bson:"created"`
		Notification models.Notification         `bson:"notification,omitempty"`
		Receiver     string                      `bson:"receiver"`
		Channel      models.Channel              `bson:"channel"`
		Status       models.TransmissionStatus   `bson:"status"`
		ResendCount  int                         `bson:"resendcount"`
		Records      []models.TransmissionRecord `bson:"records,omitempty"`
	})

	bsonErr := raw.Unmarshal(decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the non-DBRef fields
	mt.ID = decoded.ID
	mt.Created = decoded.Created
	mt.Modified = decoded.Modified
	mt.Notification = decoded.Notification
	mt.Receiver = decoded.Receiver
	mt.Channel = decoded.Channel
	mt.Status = decoded.Status
	mt.ResendCount = decoded.ResendCount
	mt.Records = decoded.Records

	return nil
}
