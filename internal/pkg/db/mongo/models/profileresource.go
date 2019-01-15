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

type ProfileResource struct {
	Name string              `bson:"name"`
	Get  []ResourceOperation `bson:"get"`
	Set  []ResourceOperation `bson:"set"`
}

func (p *ProfileResource) ToContract() (c contract.ProfileResource) {
	c.Name = p.Name
	for _, ro := range p.Get {
		c.Get = append(c.Get, ro.ToContract())
	}

	for _, ro := range p.Set {
		c.Set = append(c.Set, ro.ToContract())
	}

	return
}

func (p *ProfileResource) FromContract(c contract.ProfileResource) {
	p.Name = c.Name
	for _, ro := range c.Get {
		var get ResourceOperation
		get.FromContract(ro)
		p.Get = append(p.Get, get)
	}

	for _, ro := range c.Set {
		var set ResourceOperation
		set.FromContract(ro)
		p.Set = append(p.Set, set)
	}
}
