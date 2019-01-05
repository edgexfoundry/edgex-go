/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

import contract "github.com/edgexfoundry/edgex-go/pkg/models"

type DescribedObject struct {
	Created  int64 `bson:"created"`
	Modified int64 `bson:"modified"`
	Origin   int64 `bson:"origin"`

	Description string `bson:"description"`
}

func (do *DescribedObject) ToContract() contract.DescribedObject {
	var to contract.DescribedObject

	to.Created = do.Created
	to.Modified = do.Modified
	to.Origin = do.Origin
	to.Description = do.Description

	return to
}

func (do *DescribedObject) FromContract(from contract.DescribedObject) {
	do.Created = from.Created
	do.Modified = from.Modified
	do.Origin = from.Origin
	do.Description = from.Description
}
