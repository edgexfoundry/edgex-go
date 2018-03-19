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
		{"value", "-10.10", "-20", "20", false, true},

		{"novalue", "", "-20", "20", true, false},
		{"not_float", "data", "-20", "20", true, false},

		{"minmaxEmpty", "-10.10", "", "", false, true},

		{"min_lower", "-30", "-20", "20", true, true},
		{"max_higher", "4000", "-20", "20", true, true},

		{"notvalidmin", "-10.10", "true", "20", true, true},
		{"notvalidmax", "-10.10", "", "data", true, true},

		{"notvalidminmax", "-10.10", "true", "false", true, true},

		{"onlymin", "-10.10", "-20", "", false, true},
		{"onlymax", "-10.10", "", "20", false, true},

		{"onlymin_lower", "-110.1", "-20", "", true, true},
		{"onlymax_higher", "110.2", "", "20.09", true, true},
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
		{"value", "-10", "-20", "20", false, true},
		{"novalue", "", "", "", true, false},
		{"no_integer", "data", "-20", "20", true, false},

		{"minmaxEmpty", "-10", "", "", false, true},

		{"min_lower", "-30", "-20", "20", true, true},
		{"max_higher", "4000", "-20", "20", true, true},

		{"onlymin", "-10", "-20", "", false, true},
		{"onlymax", "-10", "", "20", false, true},

		{"notvalidmin", "-10", "true", "20", true, true},
		{"notvalidmax", "-10", "", "data", true, true},

		{"notvalidminmax", "-10", "true", "false", true, true},

		{"onlymin_lower", "-110", "-20", "", true, true},
		{"onlymax_higher", "110", "", "20", true, true},
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
		value string
		err   bool
	}{
		{"", true},
		{"B", false},
		{"b", true},
		{"P", true},

		{"F", false},
		{"f", true},
		{"P", true},

		{"I", false},
		{"i", true},
		{"P", true},

		{"S", false},
		{"s", true},
		{"P", true},

		{"J", false},
		{"j", true},
		{"P", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			tvd := models.ValueDescriptor{Type: tt.value}
			_, err := isValidValueDescriptor_private(tvd, models.Event{})
			if err != nil {
				if !tt.err {
					t.Errorf("There should not be an error: %v", err)
				}
			}
		})

	}
}
