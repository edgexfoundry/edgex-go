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
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
)

type ResourceOperation struct {
	Index     string            `bson:"index"`
	Operation string            `bson:"operation"`
	Object    string            `bson:"object"`
	Property  string            `bson:"property"`
	Parameter string            `bson:"parameter"`
	Resource  string            `bson:"resource"`
	Secondary []string          `bson:"secondary"`
	Mappings  map[string]string `bson:"mappings"`
}

func (r *ResourceOperation) ToContract() (c contract.ResourceOperation) {
	c.Index = r.Index
	c.Operation = r.Operation
	c.Object = r.Object
	c.Property = r.Property
	c.Parameter = r.Parameter
	c.Resource = r.Resource
	c.Secondary = r.Secondary
	c.Mappings = r.Mappings

	return
}

func (r *ResourceOperation) FromContract(from contract.ResourceOperation) {
	r.Index = from.Index
	r.Operation = from.Operation
	r.Object = from.Object
	r.Property = from.Property
	r.Parameter = from.Parameter
	r.Resource = from.Resource
	r.Secondary = from.Secondary
	r.Mappings = from.Mappings
}
