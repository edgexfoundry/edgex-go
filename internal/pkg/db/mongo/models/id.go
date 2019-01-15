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

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
)

func fromContractId(id string) (bson.ObjectId, string, error) {
	// In this first case, ID is empty so this must be an add.
	// Generate new BSON/UUIDs
	if id == "" {
		return bson.NewObjectId(), uuid.New().String(), nil
	}

	// In this case, we're dealing with an existing id
	if !bson.IsObjectIdHex(id) {
		// Id is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return "", "", db.ErrInvalidObjectId
		}
		return "", id, nil
	}

	// ID of pre-existing event is a BSON ID. We will query using the BSON ID.
	return bson.ObjectIdHex(id), "", nil
}

func toContractId(id bson.ObjectId, uuid string) string {
	if uuid != "" {
		return uuid
	}

	return id.Hex()
}
