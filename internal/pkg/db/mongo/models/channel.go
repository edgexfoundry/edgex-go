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

import contract "github.com/edgexfoundry/edgex-go/pkg/models"

type Channel struct {
	Type          contract.ChannelType `bson:"type,omitempty"`
	MailAddresses []string             `bson:"mailAddresses,omitempty"`
	Url           string               `bson:"url,omitempty"`
}

func (channel *Channel) ToContract() (c contract.Channel) {
	c.Type = channel.Type
	c.MailAddresses = channel.MailAddresses
	c.Url = channel.Url
	return
}

func (channel *Channel) FromContract(from contract.Channel) {
	channel.Type = from.Type
	channel.MailAddresses = from.MailAddresses
	channel.Url = from.Url
}
