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

/*
 * A subscription for notification alerts
 *
 *
 * Subscription struct
 */
type Subscription struct {
	Created              int64                            `bson:"created"`
	Modified             int64                            `bson:"modified"`
	Origin               int64                            `bson:"origin"`
	Id                   bson.ObjectId                    `bson:"_id,omitempty"`
	Uuid                 string                           `bson:"uuid,omitempty"`
	Slug                 string                           `bson:"slug"`
	Receiver             string                           `bson:"receiver"`
	Description          string                           `bson:"description"`
	SubscribedCategories []contract.NotificationsCategory `bson:"subscribedCategories,omitempty"`
	SubscribedLabels     []string                         `bson:"subscribedLabels,omitempty"`
	Channels             []Channel                        `bson:"channels,omitempty"`
}

func (s *Subscription) ToContract() (c contract.Subscription) {
	id := s.Uuid
	if id == "" {
		id = s.Id.Hex()
	}

	c.ID = id
	c.Created = s.Created
	c.Modified = s.Modified
	c.Origin = s.Origin
	c.Slug = s.Slug
	c.Receiver = s.Receiver
	c.Description = s.Description
	c.SubscribedCategories = s.SubscribedCategories
	c.SubscribedLabels = s.SubscribedLabels

	for _, channel := range s.Channels {
		c.Channels = append(c.Channels, channel.ToContract())
	}

	return
}

func (s *Subscription) FromContract(from contract.Subscription) (id string, err error) {
	s.Id, s.Uuid, err = fromContractId(from.ID)
	if err != nil {
		return
	}

	s.Created = from.Created
	s.Modified = from.Modified
	s.Origin = from.Origin
	s.Slug = from.Slug
	s.Receiver = from.Receiver
	s.Description = from.Description
	s.SubscribedCategories = from.SubscribedCategories
	s.SubscribedLabels = from.SubscribedLabels

	for _, channel := range from.Channels {
		var model Channel
		model.FromContract(channel)
		s.Channels = append(s.Channels, model)
	}

	id = toContractId(s.Id, s.Uuid)
	return
}

func (s *Subscription) TimestampForUpdate() {
	s.Modified = db.MakeTimestamp()
}

func (s *Subscription) TimestampForAdd() {
	s.TimestampForUpdate()
	s.Created = s.Modified
}
