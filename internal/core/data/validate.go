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

	models "github.com/edgexfoundry/edgex-go/pkg/models"
)

func isValidValueDescriptor(vd models.ValueDescriptor, reading models.Reading) (bool, error) {
	switch vd.Type {
	case "B": // boolean
		return validBoolean(reading)
	case "F": // floating point
		return validFloat(reading, vd)
	case "I": // integer
		return validInteger(reading, vd)
	case "S": // string or character data
		return validString(reading)
	case "J": // JSON data
		return validJSON(reading)
	default:
		return false, fmt.Errorf("Unknown type")
	}
}

func validBoolean(reading models.Reading) (bool, error) {
	_, err := strconv.ParseBool(reading.Value)
	if err != nil {
		return false, err
	}
	return true, nil
}

func validFloat(reading models.Reading, vd models.ValueDescriptor) (bool, error) {
	value, err := strconv.ParseFloat(reading.Value, 64)
	if err != nil {
		return false, err
	}

	if (vd.Max != nil) && (vd.Max != "") {
		max, err := strconv.ParseFloat(vd.Max.(string), 64)
		if err != nil {
			return false, err
		}

		if value > max {
			return false, fmt.Errorf("Value is over the limits")
		}
	}

	if (vd.Min != nil) && (vd.Min != "") {
		min, err := strconv.ParseFloat(vd.Min.(string), 64)
		if err != nil {
			return false, err
		}

		if value < min {
			return false, fmt.Errorf("Value is under the limits")
		}
	}

	return true, nil
}

func validInteger(reading models.Reading, vd models.ValueDescriptor) (bool, error) {
	value, err := strconv.ParseInt(reading.Value, 10, 64)
	if err != nil {
		return false, err
	}

	if (vd.Max != nil) && (vd.Max != "") {
		max, err := strconv.ParseInt(vd.Max.(string), 10, 64)
		if err != nil {
			return false, err
		}

		if value > max {
			return false, fmt.Errorf("Value is over the limits")
		}
	}

	if (vd.Min != nil) && (vd.Min != "") {
		min, err := strconv.ParseInt(vd.Min.(string), 10, 64)
		if err != nil {
			return false, err
		}

		if value < min {
			return false, fmt.Errorf("Value is under the limits")
		}
	}

	return true, nil
}

func validString(reading models.Reading) (bool, error) {
	if reading.Value == "" {
		return false, fmt.Errorf("Value is empty")
	}

	return true, nil
}

func validJSON(reading models.Reading) (bool, error) {
	var js interface{}
	err := json.Unmarshal([]byte(reading.Value), &js)
	if err != nil {
		return false, err
	}
	return true, nil
}
