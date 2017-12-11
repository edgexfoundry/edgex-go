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
 * @author: Ryan Comer & Spencer Bull, Dell
 * @version: 0.5.0
 *******************************************************************************/

package enums

/*
This file is the AdminState enum for EdgeX
Values are locked or unlocked
 */

type AdminStateType uint8

// The values of the enum - iota increments the value of each const
const(
	LOCKED AdminStateType = iota
	UNLOCKED
)

var adminStateStringArray = [...]string{"LOCKED", "UNLOCKED"}	// Used for String function

/*
String() func for formatting
 */
func (a AdminStateType) String() string{
	return adminStateStringArray[a]
}
