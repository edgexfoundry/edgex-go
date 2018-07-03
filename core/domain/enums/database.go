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
package enums

import (
	"errors"
)

type DATABASE int

const (
	INVALID DATABASE = iota
	MONGODB
	MEMORYDB
)

const (
	invalidStr = "invalid"
	MongoStr   = "mongodb"
	MemoryStr  = "memorydb"
)

// DATABASEArr : Add in order declared in Struct for string value
var databaseArr = [...]string{invalidStr, MongoStr, MemoryStr}

func (db DATABASE) String() string {
	if db >= INVALID && db <= MEMORYDB {
		return databaseArr[db]
	}
	return invalidStr
}

// GetDatabaseType : Return enum valude of the Database Type
func GetDatabaseType(db string) (DATABASE, error) {
	if MongoStr == db {
		return MONGODB, nil
	} else if MemoryStr == db {
		return MEMORYDB, nil
	} else {
		return INVALID, errors.New("Undefined Database Type")
	}
}
