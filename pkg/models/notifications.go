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

type Notification struct {
	BaseObject  `bson:",inline"`
	ID          bson.ObjectId         `json:"id" bson:"_id,omitempty"`
	Slug        string                `bson:"slug" json:"slug,omitempty"`
	Sender      string                `bson:"sender" json:"sender,omitempty"`
	Category    NotificationsCategory `bson:"category" json:"category,omitempty"`
	Severity    NotificationsSeverity `bson:"severity" json:"severity,omitempty"`
	Content     string                `bson:"content" json:"content,omitempty"`
	Description string                `bson:"description" json:"description,omitempty"`
	Status      NotificationsStatus   `bson:"status" json:"status,omitempty"`
	Labels      []string              `bson:"labels,omitempty" json:"labels,omitempty"`
	ContentType string                `bson:"contenttype" json:"contenttype,omitempty"`
}

func (n Notification) MarshalJSON() ([]byte, error) {
	test := struct {
		BaseObject
		ID          *bson.ObjectId        `json:"id"`
		Slug        *string               `json:"slug,omitempty,omitempty"`
		Sender      *string               `json:"sender,omitempty"`
		Category    NotificationsCategory `json:"category,omitempty"`
		Severity    NotificationsSeverity `json:"severity,omitempty"`
		Content     *string               `json:"content,omitempty"`
		Description *string               `json:"description,omitempty"`
		Status      NotificationsStatus   `json:"status,omitempty"`
		Labels      []string              `json:"labels,omitempty"`
		ContentType *string               `json:"contenttype,omitempty"`
	}{
		BaseObject: n.BaseObject,
		Category:   n.Category,
		Severity:   n.Severity,
		Status:     n.Status,
		Labels:     n.Labels,
	}

	if n.ID != "" {
		test.ID = &n.ID
	}
	if n.Slug != "" {
		test.Slug = &n.Slug
	}
	if n.Sender != "" {
		test.Sender = &n.Sender
	}
	if n.Content != "" {
		test.Content = &n.Content
	}
	if n.Description != "" {
		test.Description = &n.Description
	}
	if n.ContentType != "" {
		test.ContentType = &n.ContentType
	}
	return json.Marshal(test)
}

/*
 * To String function for Notification Struct
 */
func (n Notification) String() string {
	out, err := json.Marshal(n)
	if err != nil {
		return err.Error()
	}
	return string(out)
}
