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

func isValidValueDescriptor(reading models.Reading, ev models.Event) (bool, error) {
	vd, _ := dbc.ValueDescriptorByName(reading.Name)

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

func validBoolean(reading models.Event) (bool, error) {

	for i := range reading.Readings {
		_, err := strconv.ParseBool(reading.Readings[i].Value)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func validFloat(reading models.Event, vd models.ValueDescriptor) (bool, error) {

	if vd.Min == nil || vd.Max == nil {
		return true, nil
	}

	min, err := strconv.ParseFloat(vd.Min.(string), 64)
	if err != nil {
		return false, err
	}

	max, err := strconv.ParseFloat(vd.Max.(string), 64)
	if err != nil {
		return false, err
	}

	for i := range reading.Readings {
		value, err := strconv.ParseFloat(reading.Readings[i].Value, 64)
		if err != nil {
			return false, err
		}
		if value > max || value < min {
			return false, fmt.Errorf("Value has exceed the limits")
		}

	}
	return true, nil
}

func validInteger(reading models.Event, vd models.ValueDescriptor) (bool, error) {
	if vd.Min == nil || vd.Max == nil {
		return true, nil
	}

	min, err := strconv.ParseInt(vd.Min.(string), 10, 64)
	if err != nil {
		return false, err
	}

	max, err := strconv.ParseInt(vd.Max.(string), 10, 64)
	if err != nil {
		return false, err
	}

	for i := range reading.Readings {
		value, err := strconv.ParseInt(reading.Readings[i].Value, 10, 64)
		if err != nil {
			return false, err
		}

		if value > max || value < min {
			return false, fmt.Errorf("Value has exceed the limits")
		}

	}
	return true, nil
}

func validString(reading models.Event) (bool, error) {
	for i := range reading.Readings {
		if reading.Readings[i].Value == "" {
			return false, fmt.Errorf("Value is empty")
		}
	}

	return true, nil
}
func validJSON(reading models.Event) (bool, error) {
	var js interface{}
	for i := range reading.Readings {
		err := json.Unmarshal([]byte(reading.Readings[i].Value), &js)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}
