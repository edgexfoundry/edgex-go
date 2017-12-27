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
 *
 * @microservice: core-domain-go library
 * @author: Spencer Bull, Dell
 * @version: 0.5.0
 *******************************************************************************/
package enums

type DATABASE uint8

const (
	MONGODB DATABASE = iota
	MYSQL
)

// Add in order declared in Struct for string value
var DATABASEArr = [...]string{"mongodb", "mysql"}

func (db DATABASE) String() string {
	return DATABASEArr[db]
}
