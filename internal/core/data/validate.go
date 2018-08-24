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

func isValidValueDescriptor(vd models.ValueDescriptor, reading models.Reading) error {
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
		return fmt.Errorf("Unknown type")
	}
}

func validBoolean(reading models.Reading) error {
	_, err := strconv.ParseBool(reading.Value)
	return err
}

func validFloat(reading models.Reading, vd models.ValueDescriptor) error {
	value, err := strconv.ParseFloat(reading.Value, 64)
	if err != nil {
		return err
	}

	if (vd.Max != nil) && (vd.Max != "") {
		max, err := strconv.ParseFloat(vd.Max.(string), 64)
		if err != nil {
			return err
		}

		if value > max {
			return fmt.Errorf("Value is over the limits")
		}
	}

	if (vd.Min != nil) && (vd.Min != "") {
		min, err := strconv.ParseFloat(vd.Min.(string), 64)
		if err != nil {
			return err
		}

		if value < min {
			return fmt.Errorf("Value is under the limits")
		}
	}

	return nil
}

func validInteger(reading models.Reading, vd models.ValueDescriptor) error {
	value, err := strconv.ParseInt(reading.Value, 10, 64)
	if err != nil {
		return err
	}

	if (vd.Max != nil) && (vd.Max != "") {
		max, err := strconv.ParseInt(vd.Max.(string), 10, 64)
		if err != nil {
			return err
		}

		if value > max {
			return fmt.Errorf("Value is over the limits")
		}
	}

	if (vd.Min != nil) && (vd.Min != "") {
		min, err := strconv.ParseInt(vd.Min.(string), 10, 64)
		if err != nil {
			return err
		}

		if value < min {
			return fmt.Errorf("Value is under the limits")
		}
	}

	return nil
}

func validString(reading models.Reading) error {
	if reading.Value == "" {
		return fmt.Errorf("Value is empty")
	}

	return nil
}

func validJSON(reading models.Reading) error {
	var js interface{}
	return json.Unmarshal([]byte(reading.Value), &js)
}
