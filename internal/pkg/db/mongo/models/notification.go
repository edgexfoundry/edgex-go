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

type Notification struct {
	Created     int64                          `bson:"created"`
	Modified    int64                          `bson:"modified"`
	Origin      int64                          `bson:"origin"`
	Id          bson.ObjectId                  `bson:"_id,omitempty"`
	Uuid        string                         `bson:"uuid,omitempty"`
	Slug        string                         `bson:"slug"`
	Sender      string                         `bson:"sender"`
	Category    contract.NotificationsCategory `bson:"category"`
	Severity    contract.NotificationsSeverity `bson:"severity"`
	Content     string                         `bson:"content"`
	Description string                         `bson:"description"`
	Status      contract.NotificationsStatus   `bson:"status"`
	Labels      []string                       `bson:"labels,omitempty"`
	ContentType string                         `bson:"contenttype"`
}

func (n *Notification) ToContract() (c contract.Notification) {
	id := n.Uuid
	if id == "" {
		id = n.Id.Hex()
	}

	c.ID = id
	c.Created = n.Created
	c.Modified = n.Modified
	c.Origin = n.Origin
	c.Slug = n.Slug
	c.Sender = n.Sender
	c.Category = n.Category
	c.Severity = n.Severity
	c.Content = n.Content
	c.Description = n.Description
	c.Status = n.Status
	c.Labels = n.Labels
	c.ContentType = n.ContentType

	return
}

func (n *Notification) FromContract(from contract.Notification) (id string, err error) {
	n.Id, n.Uuid, err = fromContractId(from.ID)
	if err != nil {
		return
	}

	n.Created = from.Created
	n.Modified = from.Modified
	n.Origin = from.Origin
	n.Slug = from.Slug
	n.Sender = from.Sender
	n.Category = from.Category
	n.Severity = from.Severity
	n.Content = from.Content
	n.Description = from.Description
	n.Status = from.Status
	n.Labels = from.Labels
	n.ContentType = from.ContentType

	id = toContractId(n.Id, n.Uuid)
	return
}

func (n *Notification) TimestampForUpdate() {
	n.Modified = db.MakeTimestamp()
}

func (n *Notification) TimestampForAdd() {
	n.TimestampForUpdate()
	n.Created = n.Modified
}
