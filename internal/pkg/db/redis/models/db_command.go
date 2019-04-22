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

type Adder interface {
	Add() []DbCommand
}

type Remover interface {
	Remove() []DbCommand
}

// DbCommand is used to represent information about a specific Redis command
// If this strategy works, this type is called "DbCommand" in order to differentiate
// from the EdgeX model Command
type DbCommand struct {
	Command string //The actual Redis API command
	Hash    string //Indicates which hash index the command should be applied toward
	Key     string //Indicates the key that will be targeted in the given Hash
	Value   string //If applicable the value to be assigned to the key
	Rank    int64  //If applicable, the rank to be assigned to a given key
}
