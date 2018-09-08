/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
package data

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func TestValidBoolean(t *testing.T) {
	var tests = []struct {
		name  string
		value string
		err   bool
	}{
		{"false, nil", "false", false},
		{"true, nil", "true", false},
		{"True", "True", false},
		{"TRUE", "TRUE", false},
		{"false", "false", false},

		{"dummy, not nil ", "dummy", true},
		{"void", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reading = models.Reading{Value: tt.value}
			err := validBoolean(reading)
			if err == nil {
				if tt.err {
					t.Errorf("There should be an error: %v", err)
				}
			} else {
				if !tt.err {
					t.Errorf("There should not be an error: %v", err)
				}
			}
		})

	}
}

func TestValidFloat(t *testing.T) {

	var tests = []struct {
		name  string
		value string
		min   string
		max   string
		err   bool
	}{
		{"value", "-10.10", "-20", "20", false},

		{"novalue", "", "-20", "20", true},
		{"not_float", "data", "-20", "20", true},

		{"minmaxEmpty", "-10.10", "", "", false},

		{"min_lower", "-30", "-20", "20", true},
		{"max_higher", "4000", "-20", "20", true},

		{"notvalidmin", "-10.10", "true", "20", true},
		{"notvalidmax", "-10.10", "", "data", true},

		{"notvalidminmax", "-10.10", "true", "false", true},

		{"onlymin", "-10.10", "-20", "", false},
		{"onlymax", "-10.10", "", "20", false},

		{"onlymin_lower", "-110.1", "-20", "", true},
		{"onlymax_higher", "110.2", "", "20.09", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reading = models.Reading{Value: tt.value}
			tvd := models.ValueDescriptor{Min: tt.min, Max: tt.max}
			err := validFloat(reading, tvd)
			if err == nil {
				if tt.err {
					t.Errorf("There should be an error: %v", err)
				}
			} else {
				if !tt.err {
					t.Errorf("There should not be an error: %v", err)
				}
			}
		})

	}
}

func TestValidInteger(t *testing.T) {

	var tests = []struct {
		name  string
		value string
		min   string
		max   string
		err   bool
	}{
		{"value", "-10", "-20", "20", false},
		{"novalue", "", "", "", true},
		{"no_integer", "data", "-20", "20", true},

		{"minmaxEmpty", "-10", "", "", false},

		{"min_lower", "-30", "-20", "20", true},
		{"max_higher", "4000", "-20", "20", true},

		{"onlymin", "-10", "-20", "", false},
		{"onlymax", "-10", "", "20", false},

		{"notvalidmin", "-10", "true", "20", true},
		{"notvalidmax", "-10", "", "data", true},

		{"notvalidminmax", "-10", "true", "false", true},

		{"onlymin_lower", "-110", "-20", "", true},
		{"onlymax_higher", "110", "", "20", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reading = models.Reading{Value: tt.value}
			tvd := models.ValueDescriptor{Min: tt.min, Max: tt.max}
			err := validInteger(reading, tvd)
			if err == nil {
				if tt.err {
					t.Errorf("There should be an error: %v", err)
				}
			} else {
				if !tt.err {
					t.Errorf("There should not be an error: %v", err)
				}
			}
		})

	}
}

func TestValidString(t *testing.T) {

	var tests = []struct {
		name  string
		value string
		err   bool
	}{
		{"empty", "", true},
		{"valid", "test string", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reading = models.Reading{Value: tt.value}
			err := validString(reading)
			if err == nil {
				if tt.err {
					t.Errorf("There should be an error: %v", err)
				}
			} else {
				if !tt.err {
					t.Errorf("There should not be an error: %v", err)
				}
			}
		})

	}
}

func TestValidJson(t *testing.T) {

	var tests = []struct {
		name  string
		value string
		err   bool
	}{
		{"empty", "", true},
		{"valid", "{\"test\": \"string\"}", false},
		{"novalid", "test string", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reading = models.Reading{Value: tt.value}
			err := validJSON(reading)
			if err == nil {
				if tt.err {
					t.Errorf("There should be an error: %v", err)
				}
			} else {
				if !tt.err {
					t.Errorf("There should not be an error: %v", err)
				}
			}
		})

	}
}

func TestIsValidValueDescriptor_private(t *testing.T) {

	var tests = []struct {
		tvd   string
		value string
		err   bool
	}{
		{"", "", true},
		{"B", "true", false},
		{"b", "", true},
		{"P", "", true},

		{"F", "-10.5", false},
		{"f", "", true},
		{"P", "", true},

		{"I", "-23", false},
		{"i", "", true},
		{"P", "", true},

		{"S", "test string", false},
		{"s", "", true},
		{"P", "", true},

		{"J", "{\"test\": \"string\"}", false},
		{"j", "", true},
		{"P", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			tvd := models.ValueDescriptor{Type: tt.tvd}
			var reading = models.Reading{Value: tt.value}
			err := isValidValueDescriptor(tvd, reading)
			if err == nil {
				if tt.err {
					t.Errorf("There should be an error: %v", err)
				}
			} else {
				if !tt.err {
					t.Errorf("There should not be an error: %v", err)
				}
			}
		})
	}
}
