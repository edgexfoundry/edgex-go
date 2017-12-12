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

type ResourceOperation struct {
	Index 		string			`bson:"index" json:"index"`
	Operation 	string			`bson:"operation" json:"operation"`
	Object		string			`bson:"object" json:"object"`
	Property	string			`bson:"property" json:"property"`
	Parameter	string 			`bson:"parameter" json:"parameter"`
	Resource	string 			`bson:"resource" json:"resource"`
	Secondary	[]string		`bson:"secondary" json:"secondary"`
	Mappings	map[string]string 	`bson:"mappings" json:"mappings"`
}

// Custom marshaling to make empty strings null
func (ro ResourceOperation) MarshalJSON()([]byte, error){
	test := struct{
		Index 		*string			`json:"index"`
		Operation 	*string			`json:"operation"`
		Object		*string			`json:"object"`
		Property	*string			`json:"property"`
		Parameter	*string 			`json:"parameter"`
		Resource	*string 			`json:"resource"`
		Secondary	[]string		`json:"secondary"`
		Mappings	map[string]string 	`json:"mappings"`
	}{
		Secondary : ro.Secondary,
		Mappings : ro.Mappings,
	}
	
	// Empty strings are null
	if ro.Index != "" {test.Index = &ro.Index}
	if ro.Operation != "" {test.Operation = &ro.Operation}
	if ro.Object != "" {test.Object = &ro.Object}
	if ro.Property != "" {test.Property = &ro.Property}
	if ro.Parameter != "" {test.Parameter = &ro.Parameter}
	if ro.Resource != "" {test.Resource = &ro.Resource}
	
	return json.Marshal(test)
}

/*
 * To String function for ResourceOperation
 */
func (ro ResourceOperation) String() string {
	out, err := json.Marshal(ro)
	if err != nil {
		return err.Error()
	}
	return string(out)
}