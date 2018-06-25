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
type MongoNotification struct {
	models.Notification
}

// Custom marshaling into mongo
func (mn MongoNotification) GetBSON() (interface{}, error) {

	return struct {
		ID          bson.ObjectId                `bson:"_id,omitempty"`
		Modified    int64                        `bson:"modified"`
		Created     int64                        `bson:"created"`
		Slug        string                       `bson:"slug"`
		Sender      string                       `bson:"sender"`
		Category    models.NotificationsCategory `bson:"category"`
		Severity    models.NotificationsSeverity `bson:"severity"`
		Content     string                       `bson:"content"`
		Description string                       `bson:"description"`
		Status      models.NotificationsStatus   `bson:"status"`
		Labels      []string                     `bson:"labels,omitempty"`
		ContentType string                       `bson:"contenttype"`
	}{
		ID:          mn.ID,
		Created:     mn.Created,
		Modified:    mn.Modified,
		Slug:        mn.Slug,
		Sender:      mn.Sender,
		Category:    mn.Category,
		Severity:    mn.Severity,
		Content:     mn.Content,
		Description: mn.Description,
		Status:      mn.Status,
		Labels:      mn.Labels,
		ContentType: mn.ContentType,
	}, nil
}

// Custom unmarshaling out of mongo
func (mn *MongoNotification) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		ID          bson.ObjectId                `bson:"_id,omitempty"`
		Modified    int64                        `bson:"modified"`
		Created     int64                        `bson:"created"`
		Slug        string                       `bson:"slug"`
		Sender      string                       `bson:"sender"`
		Category    models.NotificationsCategory `bson:"category"`
		Severity    models.NotificationsSeverity `bson:"severity"`
		Content     string                       `bson:"content"`
		Description string                       `bson:"description"`
		Status      models.NotificationsStatus   `bson:"status"`
		Labels      []string                     `bson:"labels,omitempty"`
		ContentType string                       `bson:"contenttype"`
	})

	bsonErr := raw.Unmarshal(decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the non-DBRef fields
	mn.ID = decoded.ID
	mn.Created = decoded.Created
	mn.Modified = decoded.Modified
	mn.Slug = decoded.Slug
	mn.Sender = decoded.Sender
	mn.Category = decoded.Category
	mn.Severity = decoded.Severity
	mn.Content = decoded.Content
	mn.Description = decoded.Description
	mn.Status = decoded.Status
	mn.Labels = decoded.Labels
	mn.ContentType = decoded.ContentType

	return nil
}
