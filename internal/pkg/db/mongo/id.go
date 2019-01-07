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
 *
 *******************************************************************************/

package mongo

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
)

func idToQueryParameters(id string) (name string, value interface{}, err error) {
	if !bson.IsObjectIdHex(id) {
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return "", "", db.ErrInvalidObjectId
		}
		name = "uuid"
		value = id
	} else {
		name = "_id"
		value = bson.ObjectIdHex(id)
	}
	return
}

func idToBsonM(id string) (q bson.M, err error) {
	var name string
	var value interface{}
	name, value, err = idToQueryParameters(id)
	if err != nil {
		return
	}
	q = bson.M{name: value}
	return
}
