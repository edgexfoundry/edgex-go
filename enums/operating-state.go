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
This file is the Operating state enum for EdgeX
Values are enabled or disabled
 */

type OperatingStateType uint8

const(
	ENABLED OperatingStateType = iota
	DISABLED
)

var operatingStateStringArray = [...]string{"ENABLED", "DISABLED"}	// Used for String() function

/*
String function for formatting
 */
func (o OperatingStateType) String()string{
	return operatingStateStringArray[o]
}
