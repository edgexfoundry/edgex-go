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
type MongoSubscription struct {
	models.Subscription
}

// Custom marshaling into mongo
func (ms MongoSubscription) GetBSON() (interface{}, error) {

	return struct {
		ID                   bson.ObjectId                  `bson:"_id,omitempty"`
		Modified             int64                          `bson:"modified"`
		Created              int64                          `bson:"created"`
		Slug                 string                         `bson:"slug"`
		Receiver             string                         `bson:"receiver"`
		Description          string                         `bson:"description"`
		SubscribedCategories []models.NotificationsCategory `bson:"subscribedcategories,omitempty"`
		SubscribedLabels     []string                       `bson:"subscribedlabels,omitempty"`
		Channels             []models.Channel               `bson:"channels,omitempty"`
	}{
		ID:                   ms.ID,
		Created:              ms.Created,
		Modified:             ms.Modified,
		Slug:                 ms.Slug,
		Receiver:             ms.Receiver,
		Description:          ms.Description,
		SubscribedCategories: ms.SubscribedCategories,
		SubscribedLabels:     ms.SubscribedLabels,
		Channels:             ms.Channels,
	}, nil
}

// Custom unmarshaling out of mongo
func (ms *MongoSubscription) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		ID                   bson.ObjectId                  `bson:"_id,omitempty"`
		Modified             int64                          `bson:"modified"`
		Created              int64                          `bson:"created"`
		Slug                 string                         `bson:"slug"`
		Receiver             string                         `bson:"receiver"`
		Description          string                         `bson:"description"`
		SubscribedCategories []models.NotificationsCategory `bson:"subscribedcategories,omitempty"`
		SubscribedLabels     []string                       `bson:"subscribedlabels,omitempty"`
		Channels             []models.Channel               `bson:"channels,omitempty"`
	})

	bsonErr := raw.Unmarshal(decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the non-DBRef fields
	ms.ID = decoded.ID
	ms.Created = decoded.Created
	ms.Modified = decoded.Modified
	ms.Slug = decoded.Slug
	ms.Receiver = decoded.Receiver
	ms.Description = decoded.Description
	ms.SubscribedCategories = decoded.SubscribedCategories
	ms.SubscribedLabels = decoded.SubscribedLabels
	ms.Channels = decoded.Channels

	return nil
}
