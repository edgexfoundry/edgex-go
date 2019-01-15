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
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
)

type TransmissionRecord struct {
	Status   contract.TransmissionStatus `bson:"status"`
	Response string                      `bson:"response"`
	Sent     int64                       `bson:"sent"`
}

func (tr *TransmissionRecord) ToContract() (c contract.TransmissionRecord) {
	c.Status = tr.Status
	c.Response = tr.Response
	c.Sent = tr.Sent
	return
}

func (tr *TransmissionRecord) FromContract(from contract.TransmissionRecord) {
	tr.Status = from.Status
	tr.Response = from.Response
	tr.Sent = from.Sent
}
