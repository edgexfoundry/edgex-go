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

import contract "github.com/edgexfoundry/edgex-go/pkg/models"

type PropertyValue struct {
	Type         string `bson:"type"`         // ValueDescriptor Type of property after transformations
	ReadWrite    string `bson:"readWrite"`    // Read/Write Permissions set for this property
	Minimum      string `bson:"minimum"`      // Minimum value that can be get/set from this property
	Maximum      string `bson:"maximum"`      // Maximum value that can be get/set from this property
	DefaultValue string `bson:"defaultValue"` // Default value set to this property if no argument is passed
	Size         string `bson:"size"`         // Size of this property in its type  (i.e. bytes for numeric types, characters for string types)
	Word         string `bson:"word"`         // Word size of property used for endianness
	LSB          string `bson:"lsb"`          // Endianness setting for a property
	Mask         string `bson:"mask"`         // Mask to be applied prior to get/set of property
	Shift        string `bson:"shift"`        // Shift to be applied after masking, prior to get/set of property
	Scale        string `bson:"scale"`        // Multiplicative factor to be applied after shifting, prior to get/set of property
	Offset       string `bson:"offset"`       // Additive factor to be applied after multiplying, prior to get/set of property
	Base         string `bson:"base"`         // Base for property to be applied to, leave 0 for no power operation (i.e. base ^ property: 2 ^ 10)
	Assertion    string `bson:"assertion"`    // Required value of the property, set for checking error state.  Failing an assertion condition will mark the device with an error state
	Signed       bool   `bson:"signed"`       // Treat the property as a signed or unsigned value
	Precision    string `bson:"precision"`
}

func (pv *PropertyValue) ToContract() (c contract.PropertyValue) {
	c.Type = pv.Type
	c.ReadWrite = pv.ReadWrite
	c.Minimum = pv.Minimum
	c.Maximum = pv.Maximum
	c.DefaultValue = pv.DefaultValue
	c.Size = pv.Size
	c.Word = pv.Word
	c.LSB = pv.LSB
	c.Mask = pv.Mask
	c.Shift = pv.Shift
	c.Scale = pv.Scale
	c.Offset = pv.Offset
	c.Base = pv.Base
	c.Assertion = pv.Assertion
	c.Signed = pv.Signed
	c.Precision = pv.Precision

	return
}

func (pv *PropertyValue) FromContract(c contract.PropertyValue) {
	pv.Type = c.Type
	pv.ReadWrite = c.ReadWrite
	pv.Minimum = c.Minimum
	pv.Maximum = c.Maximum
	pv.DefaultValue = c.DefaultValue
	pv.Size = c.Size
	pv.Word = c.Word
	pv.LSB = c.LSB
	pv.Mask = c.Mask
	pv.Shift = c.Shift
	pv.Scale = c.Scale
	pv.Offset = c.Offset
	pv.Base = c.Base
	pv.Assertion = c.Assertion
	pv.Signed = c.Signed
	pv.Precision = c.Precision
}
