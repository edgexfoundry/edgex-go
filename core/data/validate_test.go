package data

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
)

func TestValidBoolean(t *testing.T) {
	var tests = []struct {
		name   string
		value  string
		err    bool
		result bool
	}{
		{"false, nil", "false", false, true},
		{"true, nil", "true", false, true},
		{"True", "True", false, true},
		{"TRUE", "TRUE", false, true},
		{"false", "false", false, true},

		{"dummy, not nil ", "dummy", true, false},
		{"void", "", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event = models.Event{}
			event.Readings = append(event.Readings, models.Reading{Value: tt.value})
			val, err := validBoolean(event)
			if err == nil {
				if tt.result != val {
					t.Errorf("expecting %v, returned %v", tt.result, val)
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
		name   string
		value  string
		min    string
		max    string
		err    bool
		result bool
	}{
		{"float", "-10.10", "-20", "20", false, true},
		{"minmaxEmpty", "-10.10", "", "", false, true},

		{"nofloat", "data", "-20", "20", true, false},

		{"min", "-30", "-20", "20", true, true},
		{"max", "4000", "-20", "20", true, true},

		{"nomin", "-10.10", "true", "20", true, true},
		{"nomax", "-10.10", "", "data", true, true},

		{"onlymin", "-10.10", "-20", "", false, true},
		{"onlymax", "-10.10", "", "20", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event = models.Event{}
			event.Readings = append(event.Readings, models.Reading{Value: tt.value})
			tvd := models.ValueDescriptor{Min: tt.min, Max: tt.max}
			val, err := validFloat(event, tvd)
			if err == nil {
				if tt.result != val {
					t.Errorf("expecting %v, returned %v", tt.result, val)
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
		name   string
		value  string
		min    string
		max    string
		err    bool
		result bool
	}{
		{"float", "-10", "-20", "20", false, true},
		{"minmaxEmpty", "-10", "", "", false, true},

		{"nofloat", "data", "-20", "20", true, false},

		{"min", "-30", "-20", "20", true, true},
		{"max", "4000", "-20", "20", true, true},

		{"nomin", "-10", "true", "20", true, true},
		{"nomax", "-10", "", "data", true, true},

		{"onlymin", "-10", "-20", "", false, true},
		{"onlymax", "-10", "", "20", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event = models.Event{}
			event.Readings = append(event.Readings, models.Reading{Value: tt.value})
			tvd := models.ValueDescriptor{Min: tt.min, Max: tt.max}
			val, err := validInteger(event, tvd)
			if err == nil {
				if tt.result != val {
					t.Errorf("expecting %v, returned %v", tt.result, val)
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
		name   string
		value  string
		err    bool
		result bool
	}{
		{"empty", "", true, true},
		{"valid", "test string", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event = models.Event{}

			event.Readings = append(event.Readings, models.Reading{Value: tt.value})
			val, err := validString(event)
			if err == nil {
				if tt.result != val {
					t.Errorf("expecting %v, returned %v", tt.result, val)
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
		name   string
		value  string
		err    bool
		result bool
	}{
		{"empty", "", true, true},
		{"valid", "{\"test\": \"string\"}", false, true},
		{"novalid", "test string", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event = models.Event{}
			event.Readings = append(event.Readings, models.Reading{Value: tt.value})
			val, err := validJSON(event)
			if err == nil {
				if tt.result != val {
					t.Errorf("expecting %v, returned %v", tt.result, val)
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
		name  string
		value string
		err   bool
	}{
		{"empty", "", true},
		{"boolean_B", "B", false},
		{"boolean_b", "b", true},
		{"boolean_p", "P", true},

		{"float_F", "F", false},
		{"float_f", "f", true},
		{"float_p", "P", true},

		{"integer_I", "I", false},
		{"integer_i", "i", true},
		{"integer_p", "P", true},

		{"string_S", "S", false},
		{"string_s", "s", true},
		{"string_p", "P", true},

		{"json_J", "J", false},
		{"json_j", "j", true},
		{"json_p", "P", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tvd := models.ValueDescriptor{Type: tt.value}
			_, err := isValidValueDescriptor_private(tvd, models.Reading{}, models.Event{})
			if err != nil {
				if !tt.err {
					t.Errorf("There should not be an error: %v", err)
				}
			}
		})

	}
}
