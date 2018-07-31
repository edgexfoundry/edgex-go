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

type Transmission struct {
	BaseObject   `bson:",inline"`
	ID           bson.ObjectId        `json:"id" bson:"_id,omitempty"`
	Notification Notification         `json:"notification" bson:"notification,omitempty"`
	Receiver     string               `bson:"receiver" json:"receiver,omitempty"`
	Channel      Channel              `bson:"channel,omitempty" json:"channel,omitempty"`
	Status       TransmissionStatus   `bson:"status" json:"status,omitempty"`
	ResendCount  int                  `bson:"resendcount" json:"resendcount"`
	Records      []TransmissionRecord `bson:"records,omitempty" json:"records,omitempty"`
}

// Custom marshaling to make empty strings null
func (t Transmission) MarshalJSON() ([]byte, error) {
	test := struct {
		BaseObject
		ID           *bson.ObjectId       `json:"id"`
		Notification Notification         `json:"notification,omitempty"`
		Receiver     *string              `json:"receiver,omitempty"`
		Channel      Channel              `json:"channel,omitempty"`
		Status       TransmissionStatus   `json:"status,omitempty"`
		ResendCount  int                  `json:"resendcount"`
		Records      []TransmissionRecord `json:"records,omitempty"`
	}{
		BaseObject:   t.BaseObject,
		Notification: t.Notification,
		Channel:      t.Channel,
		Status:       t.Status,
		ResendCount:  t.ResendCount,
		Records:      t.Records,
	}
	// Empty strings are null
	if t.Receiver != "" {
		test.Receiver = &t.Receiver
	}
	return json.Marshal(test)
}

/*
 * To String function for Transmission Struct
 */
func (t Transmission) String() string {
	out, err := json.Marshal(t)
	if err != nil {
		return err.Error()
	}
	return string(out)
}
