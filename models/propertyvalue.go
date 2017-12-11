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

import(
	"encoding/json"
)

type PropertyValue struct {
	Type 		string	`bson:"type" json:"type"`			// ValueDescriptor Type of property after transformations
	ReadWrite 	string	`bson:"readWrite" json:"readWrite" yaml:"readWrite"`		// Read/Write Permissions set for this property
	Minimum		string	`bson:"minimum" json:"minimum"`			// Minimum value that can be get/set from this property
	Maximum		string	`bson:"maximum" json:"maximum"`			// Maximum value that can be get/set from this property
	DefaultValue	string	`bson:"defaultValue" json:"defaultValue" yaml:"defaultValue"`	// Default value set to this property if no argument is passed
	Size		string	`bson:"size" json:"size"`			// Size of this property in its type  (i.e. bytes for numeric types, characters for string types)
	Word 		string 	`bson:"word" json:"word"`			// Word size of property used for endianness
	LSB		string	`bson:"lsb" json:"lsb"`				// Endianness setting for a property
	Mask		string	`bson:"mask" json:"mask"`			// Mask to be applied prior to get/set of property
	Shift		string 	`bson:"shift" json:"shift"`			// Shift to be applied after masking, prior to get/set of property
	Scale		string 	`bson:"scale" json:"scale"`			// Multiplicative factor to be applied after shifting, prior to get/set of property
	Offset		string 	`bson:"offset" json:"offset"`			// Additive factor to be applied after multiplying, prior to get/set of property
	Base		string	`bson:"base" json:"base"`			// Base for property to be applied to, leave 0 for no power operation (i.e. base ^ property: 2 ^ 10)
	Assertion	string	`bson:"assertion" json:"assertion"`		// Required value of the property, set for checking error state.  Failing an assertion condition will mark the device with an error state
	Signed		bool 	`bson:"signed" json:"signed"`			// Treat the property as a signed or unsigned value
	Precision string	`bson:"precision" json:"precision"`
}

// Custom marshaling to make empty strings null
func (pv PropertyValue) MarshalJSON()([]byte, error){
	test := struct{
		Type 		*string	`json:"type"`			// ValueDescriptor Type of property after transformations
		ReadWrite 	*string	`json:"readWrite"`		// Read/Write Permissions set for this property
		Minimum		*string	`json:"minimum"`			// Minimum value that can be get/set from this property
		Maximum		*string	`json:"maximum"`			// Maximum value that can be get/set from this property
		DefaultValue	*string	`json:"defaultValue"`	// Default value set to this property if no argument is passed
		Size		*string	`json:"size"`			// Size of this property in its type  (i.e. bytes for numeric types, characters for string types)
		Word 		*string 	`json:"word"`			// Word size of property used for endianness
		LSB		*string	`json:"lsb"`				// Endianness setting for a property
		Mask		*string	`json:"mask"`			// Mask to be applied prior to get/set of property
		Shift		*string 	`json:"shift"`			// Shift to be applied after masking, prior to get/set of property
		Scale		*string 	`json:"scale"`			// Multiplicative factor to be applied after shifting, prior to get/set of property
		Offset		*string 	`json:"offset"`			// Additive factor to be applied after multiplying, prior to get/set of property
		Base		*string	`json:"base"`			// Base for property to be applied to, leave 0 for no power operation (i.e. base ^ property: 2 ^ 10)
		Assertion	*string	`json:"assertion"`		// Required value of the property, set for checking error state.  Failing an assertion condition will mark the device with an error state
		Signed		bool 	`json:"signed"`			// Treat the property as a signed or unsigned value
		Precision *string	`json:"precision"`
	}{
		Signed : pv.Signed,
	}
	
	// Empty strings are null
	if pv.Type != "" {test.Type = &pv.Type}
	if pv.ReadWrite != "" {test.ReadWrite = &pv.ReadWrite}
	if pv.Minimum != "" {test.Minimum = &pv.Minimum}
	if pv.Maximum != "" {test.Maximum = &pv.Maximum}
	if pv.DefaultValue != "" {test.DefaultValue = &pv.DefaultValue}
	if pv.Size != "" {test.Size = &pv.Size}
	if pv.Word != "" {test.Word = &pv.Word}
	if pv.LSB != "" {test.LSB = &pv.LSB}
	if pv.Mask != "" {test.Mask = &pv.Mask}
	if pv.Shift != "" {test.Shift = &pv.Shift}
	if pv.Scale != "" {test.Scale = &pv.Scale}
	if pv.Offset != "" {test.Offset = &pv.Offset}
	if pv.Base != "" {test.Base = &pv.Base}
	if pv.Assertion != "" {test.Assertion = &pv.Assertion}
	if pv.Precision != "" {test.Precision = &pv.Precision}
	
	return json.Marshal(test)
}

/*
 * To String function for DeviceService
 */
func (pv PropertyValue) String() string {
	out, err := json.Marshal(pv)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

// Custom unmarshaling to handle default values
func (p *PropertyValue) UnmarshalJSON(data []byte) error{
	type testAlias PropertyValue
	test := testAlias{Word : "2", Mask : "0x00", Shift : "0", Scale : "1.0", Offset : "0.0", Base : "0", Signed : true}
	if err := json.Unmarshal(data, &test); err != nil{return err}
	
	// Set the default values
//	if test.Word == "" {test.Word = "2"}
//	if test.Mask == "" {test.Mask = "0x00"}
//	if test.Shift == "" {test.Shift = "0"}
//	if test.Scale == "" {test.Scale = "1.0"}
//	if test.Offset == "" {test.Offset = "0.0"}
//	if test.Base == "" {test.Base = "0"}
	
	*p = PropertyValue(test)
	
	return nil
}

// Custom YAML unmarshaling
func(p *PropertyValue) UnmarshalYAML(unmarshal func(interface{}) error) error{
	type testAlias PropertyValue
	test := testAlias{Word : "2", Mask : "0x00", Shift : "0", Scale : "1.0", Offset : "0.0", Base : "0", Signed : true}
	if err := unmarshal(&test); err != nil{return err}
	
	// Set the default values
//	if test.Word == "" {test.Word = "2"}
//	if test.Mask == "" {test.Mask = "0x00"}
//	if test.Shift == "" {test.Shift = "0"}
//	if test.Scale == "" {test.Scale = "1.0"}
//	if test.Offset == "" {test.Offset = "0.0"}
//	if test.Base == "" {test.Base = "0"}
	
	*p = PropertyValue(test)
	
	return nil
}
