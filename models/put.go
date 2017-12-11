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

package models

import "encoding/json"
/*
 * This file is the model for Put commands in EdgeX
 *
 * Put Struct
 */
type Put struct {
	Action				`bson:",inline" yaml:",inline"`
	ParameterNames []string		`bson:"parameterNames" json:"parameterNames" yaml:"parameterNames"`
}

// Custom marshaling to make empty strings null
func (p Put) MarshalJSON()([]byte, error){
	test := struct{
		Path *string `json:"path"`
		Responses []Response `json:"responses"`
		ParameterNames []string `json:"parameterNames"`
	}{}
	
	// Empty strings are null
	if p.Path != ""{test.Path = &p.Path}
	
	// Empty arrays are null
	if len(p.Responses) > 0 {test.Responses = p.Responses}
	if len(p.ParameterNames) > 0 {test.ParameterNames = p.ParameterNames}
	
	return json.Marshal(test)
}

/*
 * To String function for Put struct
 */
func (p Put) String() string {
	out, err := json.Marshal(p)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

// Append the associated value descriptors to the list
func (p *Put) AllAssociatedValueDescriptors(vdNames *map[string]string){
	for _, pn := range p.ParameterNames{
		// Only add to the map if the value descriptor is NOT there
		if _, ok := (*vdNames)[pn]; !ok{
			(*vdNames)[pn] = pn
		}
	}
}
