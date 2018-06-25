/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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
 *
 *******************************************************************************/

package models

import (
	"encoding/json"

	"gopkg.in/mgo.v2/bson"
)

/*
 * A subscription for notification alerts
 *
 *
 * Subscription struct
 */
type Subscription struct {
	BaseObject           `bson:",inline"`
	ID                   bson.ObjectId           `bson:"_id,omitempty" json:"id"`
	Slug                 string                  `bson:"slug" json:"slug,omitempty"`
	Receiver             string                  `bson:"receiver" json:"receiver,omitempty"`
	Description          string                  `bson:"description" json:"description,omitempty"`
	SubscribedCategories []NotificationsCategory `bson:"subscribedcategories,omitempty" json:"subscribedcategories,omitempty"`
	SubscribedLabels     []string                `bson:"subscribedlabels,omitempty" json:"subscribedlabels,omitempty"`
	Channels             []Channel               `bson:"channels,omitempty" json:"channels,omitempty"`
}

// Custom marshaling to make empty strings null
func (s Subscription) MarshalJSON() ([]byte, error) {
	test := struct {
		BaseObject
		ID                   *bson.ObjectId          `json:"id"`
		Slug                 *string                 `json:"slug,omitempty"`
		Receiver             *string                 `json:"receiver,omitempty"`
		Description          *string                 `json:"description,omitempty"`
		SubscribedCategories []NotificationsCategory `json:"subscribedcategories,omitempty"`
		SubscribedLabels     []string                `json:"subscribedlabels,omitempty"`
		Channels             []Channel               `json:"channels,omitempty"`
	}{
		BaseObject:           s.BaseObject,
		SubscribedCategories: s.SubscribedCategories,
		SubscribedLabels:     s.SubscribedLabels,
		Channels:             s.Channels,
	}

	if s.ID != "" {
		test.ID = &s.ID
	}
	if s.Slug != "" {
		test.Slug = &s.Slug
	}
	if s.Receiver != "" {
		test.Receiver = &s.Receiver
	}
	if s.Description != "" {
		test.Description = &s.Description
	}
	return json.Marshal(test)
}

/*
 * To String function for Notification Struct
 */
func (s Subscription) String() string {
	out, err := json.Marshal(s)
	if err != nil {
		return err.Error()
	}
	return string(out)
}
