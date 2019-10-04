/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
 *******************************************************************************/

package models

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/globalsign/mgo/bson"
)

type Response struct {
	Code           string   `bson:"code"`
	Description    string   `bson:"description"`
	ExpectedValues []string `bson:"expectedValues"`
}

type Get struct {
	Path      string     `bson:"path"`      // path used by service for action on a device or sensor
	Responses []Response `bson:"responses"` // responses from get or put requests to service
	URL       string     // url for requests from command service
}

type Put struct {
	Path           string     `bson:"path"`      // path used by service for action on a device or sensor
	Responses      []Response `bson:"responses"` // responses from get or put requests to service
	URL            string     // url for requests from command service
	ParameterNames []string   `bson:"parameterNames"`
}

type CommandProfile struct {
	Name string `bson:"name"`
	Get  Get    `bson:"get"`
	Put  Put    `bson:"put"`
}

func (c *CommandProfile) ToContract() (cmd contract.Command) {
	cmd.Get = contract.Get{}
	cmd.Get.Path = c.Get.Path
	cmd.Get.URL = c.Get.URL
	cmd.Get.Responses = []contract.Response{}
	for _, r := range c.Get.Responses {
		cmd.Get.Responses = append(cmd.Get.Responses, contract.Response{
			Code:           r.Code,
			Description:    r.Description,
			ExpectedValues: r.ExpectedValues,
		})
	}

	cmd.Put = contract.Put{}
	cmd.Put.Path = c.Put.Path
	cmd.Put.Responses = []contract.Response{}
	for _, r := range c.Put.Responses {
		cmd.Put.Responses = append(cmd.Put.Responses, contract.Response{
			Code:           r.Code,
			Description:    r.Description,
			ExpectedValues: r.ExpectedValues,
		})
	}
	cmd.Put.URL = c.Put.URL
	cmd.Put.ParameterNames = c.Put.ParameterNames

	cmd.Name = c.Name

	return
}

func (c *CommandProfile) FromContract(from contract.Command) (contractId string, err error) {
	c.Name = from.Name

	c.Get = Get{}
	c.Get.Path = from.Get.Path
	c.Get.URL = from.Get.URL
	c.Get.Responses = []Response{}
	for _, val := range from.Get.Responses {
		c.Get.Responses = append(c.Get.Responses, Response{
			Code:           val.Code,
			Description:    val.Description,
			ExpectedValues: val.ExpectedValues,
		})
	}

	c.Put = Put{}
	c.Put.Path = from.Put.Path
	c.Put.Responses = []Response{}
	for _, val := range from.Put.Responses {
		c.Put.Responses = append(c.Put.Responses, Response{
			Code:           val.Code,
			Description:    val.Description,
			ExpectedValues: val.ExpectedValues,
		})
	}
	c.Put.URL = from.Put.URL
	c.Put.ParameterNames = from.Put.ParameterNames
	return
}

func (c *Command) TimestampForUpdate() {
	c.Modified = db.MakeTimestamp()
}

func (c *Command) TimestampForAdd() {
	c.TimestampForUpdate()
	c.Created = c.Modified
}

type Command struct {
	CommandProfile `bson:",inline"`
	Id             bson.ObjectId `bson:"_id,omitempty"`
	Uuid           string        `bson:"uuid,omitempty"`
	DeviceID       string        `bson:"deviceId"`
	DeviceName     string        `bson:"deviceName"`
	Created        int64         `bson:"created"`
	Modified       int64         `bson:"modified"`
	Origin         int64         `bson:"origin"`
}

func (c *Command) ToContract() (cmd contract.Command) {
	cmd = c.CommandProfile.ToContract()
	id := c.Uuid
	if id == "" {
		id = c.Id.Hex()
	}
	cmd.Id = id
	cmd.Created = c.Created
	cmd.Modified = c.Modified
	cmd.Origin = c.Origin
	return
}

func (c *Command) FromContract(fromCommand contract.Command, fromDeviceId string, fromDeviceName string) (contractId string, err error) {
	c.Id, c.Uuid, err = fromContractId(fromCommand.Id)
	if err != nil {
		return
	}
	c.CommandProfile.FromContract(fromCommand)
	c.DeviceID = fromDeviceId
	c.DeviceName = fromDeviceName
	c.Created = fromCommand.Created
	c.Modified = fromCommand.Modified
	c.Origin = fromCommand.Origin

	contractId = toContractId(c.Id, c.Uuid)
	return
}
