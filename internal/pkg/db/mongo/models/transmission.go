/*******************************************************************************
 * Copyright 2019 Dell Technologies Inc.
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
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo/bson"
)

type TransmissionRecord struct {
	Status   contract.TransmissionStatus `bson:"status"`
	Response string                      `bson:"response"`
	Sent     int64                       `bson:"sent"`
}

type Transmission struct {
	Created      int64                       `bson:"created"`
	Modified     int64                       `bson:"modified"`
	Origin       int64                       `bson:"origin"`
	Id           bson.ObjectId               `bson:"_id,omitempty"`
	Uuid         string                      `bson:"uuid,omitempty"`
	Notification Notification                `bson:"notification,omitempty"`
	Receiver     string                      `bson:"receiver"`
	Channel      Channel                     `bson:"channel,omitempty"`
	Status       contract.TransmissionStatus `bson:"status"`
	ResendCount  int                         `bson:"resendcount"`
	Records      []TransmissionRecord        `bson:"records,omitempty"`
}

func (t *Transmission) ToContract() (c contract.Transmission) {
	id := t.Uuid
	if id == "" {
		id = t.Id.Hex()
	}

	c.ID = id
	c.Created = t.Created
	c.Modified = t.Modified
	c.Origin = t.Origin
	c.Notification = t.Notification.ToContract()
	c.Receiver = t.Receiver
	c.Channel = t.Channel.ToContract()
	c.Status = t.Status
	c.ResendCount = t.ResendCount

	for _, record := range t.Records {
		c.Records = append(c.Records, contract.TransmissionRecord{
			Status:   record.Status,
			Response: record.Response,
			Sent:     record.Sent,
		})
	}

	return
}

func (t *Transmission) FromContract(from contract.Transmission) (id string, err error) {
	t.Id, t.Uuid, err = fromContractId(from.ID)
	if err != nil {
		return
	}

	t.Created = from.Created
	t.Modified = from.Modified
	t.Origin = from.Origin
	if _, err = t.Notification.FromContract(from.Notification); err != nil {
		return
	}
	t.Receiver = from.Receiver
	t.Channel.FromContract(from.Channel)
	t.Status = from.Status
	t.ResendCount = from.ResendCount

	for _, record := range from.Records {
		t.Records = append(t.Records, TransmissionRecord{
			Status:   record.Status,
			Response: record.Response,
			Sent:     record.Sent,
		})
	}

	id = toContractId(t.Id, t.Uuid)
	return
}

func (t *Transmission) TimestampForUpdate() {
	t.Modified = db.MakeTimestamp()
}

func (t *Transmission) TimestampForAdd() {
	t.TimestampForUpdate()
	t.Created = t.Modified
}
