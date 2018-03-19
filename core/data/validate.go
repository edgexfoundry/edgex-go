//
// Copyright (c) 2018
// Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package data

import (
	"encoding/json"
	"fmt"
	"strconv"

	models "github.com/edgexfoundry/edgex-go/core/domain/models"
)

func isValidValueDescriptor_private(vd models.ValueDescriptor, ev models.Event) (bool, error) {
	switch vd.Type {
	case "B": // boolean
		return validBoolean(ev)
	case "F": // floating point
		return validFloat(ev, vd)
	case "I": // integer
		return validInteger(ev, vd)
	case "S": // string or character data
		return validString(ev)
	case "J": // JSON data
		return validJSON(ev)
	default:
		return false, fmt.Errorf("Unknown type")
	}
}

func isValidValueDescriptor(reading models.Reading, ev models.Event) (bool, error) {
	vd, _ := dbc.ValueDescriptorByName(reading.Name)
	return isValidValueDescriptor_private(vd, ev)
}

func validBoolean(ev models.Event) (bool, error) {

	for i := range ev.Readings {
		_, err := strconv.ParseBool(ev.Readings[i].Value)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func validFloat(ev models.Event, vd models.ValueDescriptor) (bool, error) {

	//check for empty limits
	maxLimit := true
	if (vd.Max == nil) || (vd.Max == "") {
		maxLimit = false
	}

	minLimit := true
	if (vd.Min == nil) || (vd.Min == "") {
		minLimit = false
	}

	var err error

	min := 0.0
	max := 0.0

	if minLimit {
		min, err = strconv.ParseFloat(vd.Min.(string), 64)
		if err != nil {
			return false, err
		}
	}

	if maxLimit {
		max, err = strconv.ParseFloat(vd.Max.(string), 64)
		if err != nil {
			return false, err
		}

	}

	for i := range ev.Readings {
		value, err := strconv.ParseFloat(ev.Readings[i].Value, 64)
		if err != nil {
			return false, err
		}

		if maxLimit {
			if value > max {
				return false, fmt.Errorf("Value is over the limits")
			}
		}

		if minLimit {
			if value < min {
				return false, fmt.Errorf("Value is under the limits")
			}
		}
	}
	return true, nil
}

func validInteger(ev models.Event, vd models.ValueDescriptor) (bool, error) {
	//check for empty limits
	maxLimit := true
	if (vd.Max == nil) || (vd.Max == "") {
		maxLimit = false
	}

	minLimit := true
	if (vd.Min == nil) || (vd.Min == "") {
		minLimit = false
	}

	bothLimits := true
	if !minLimit && !maxLimit {
		bothLimits = false
	}

	var err error

	min := int64(0)
	max := int64(0)

	if minLimit {
		min, err = strconv.ParseInt(vd.Min.(string), 10, 64)
		if err != nil {
			return false, err
		}
	}

	if maxLimit {
		max, err = strconv.ParseInt(vd.Max.(string), 10, 64)
		if err != nil {
			return false, err
		}

	}

	for i := range ev.Readings {
		value, err := strconv.ParseInt(ev.Readings[i].Value, 10, 64)
		if err != nil {
			return false, err
		}

		if !bothLimits {
			return true, nil
		}

		if maxLimit {
			if value > max {
				return false, fmt.Errorf("Value is over the limits")
			}
		}
		if minLimit {
			if value < min {
				return false, fmt.Errorf("Value is under the limits")
			}
		}
	}
	return true, nil
}

func validString(ev models.Event) (bool, error) {
	for i := range ev.Readings {
		if ev.Readings[i].Value == "" {
			return false, fmt.Errorf("Value is empty")
		}
	}

	return true, nil
}
func validJSON(ev models.Event) (bool, error) {
	var js interface{}
	for i := range ev.Readings {
		err := json.Unmarshal([]byte(ev.Readings[i].Value), &js)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}
