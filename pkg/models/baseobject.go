/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
	"encoding/json"
)

/*
TODO: Since this type is embedded throughout many contract models, we must leave the bson tags in place until
the full decoupling from Mongo is complete. This shouldn't have any impact since there isn't a direct import
to mgo/bson
*/
type BaseObject struct {
	Created  int64 `bson:"created" json:"created"`
	Modified int64 `bson:"modified" json:"modified"`
	Origin   int64 `bson:"origin" json:"origin"`
}

/*
 * String function for representing a device
 */
func (o *BaseObject) String() string {
	out, err := json.Marshal(o)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

/*
 * Compare the Created of two objects to determine given is newer
 */
func (ba *BaseObject) compareTo(i BaseObject) int {
	if i.Created > ba.Created {
		return 1
	}
	return -1
}
