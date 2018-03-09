//
// Copyright (c) 2018
// Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package data

import (
	"encoding/json"
	"strconv"

	models "github.com/edgexfoundry/edgex-go/core/domain/models"
)

func isValidValueDescriptor(reading models.Reading, ev models.Event) bool {
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
		return false
	}

	return false
}

func validBoolean(reading models.Event) bool {

	for i := range reading.Readings {
		_, err := strconv.ParseBool(reading.Readings[i].Value)
		if err != nil {
			return false
		}
	}
	return true
}

func validFloat(reading models.Event, vd models.ValueDescriptor) bool {

	if vd.Min == nil || vd.Max == nil {
		return true
	}

	min, err := strconv.ParseFloat(vd.Min.(string), 64)
	if err != nil {
		loggingClient.Error("Error: ", err.Error())
		return false
	}

	max, err := strconv.ParseFloat(vd.Max.(string), 64)
	if err != nil {
		loggingClient.Error("Error: ", err.Error())
		return false
	}

	for i := range reading.Readings {
		value, err := strconv.ParseFloat(reading.Readings[i].Value, 64)
		if err != nil {
			loggingClient.Error("Error: ", err.Error())
			return false
		}
		if value > max || value < min {
			return false
		}

	}
	return true
}

func validInteger(reading models.Event, vd models.ValueDescriptor) bool {
	if vd.Min == nil || vd.Max == nil {
		return true
	}

	min, err := strconv.ParseInt(vd.Min.(string), 10, 64)
	if err != nil {
		loggingClient.Error("Error: ", err.Error())
		return false
	}

	max, err := strconv.ParseInt(vd.Max.(string), 10, 64)
	if err != nil {
		loggingClient.Error("Error: ", err.Error())
		return false
	}

	for i := range reading.Readings {

		value, err := strconv.ParseInt(reading.Readings[i].Value, 10, 64)
		if err != nil {
			loggingClient.Error("Error: ", err.Error())
			return false
		}
		if value > max || value < min {
			return false
		}

	}
	return true
}

func validString(reading models.Event) bool {
	for i := range reading.Readings {
		if reading.Readings[i].Value == "" {
			return false
		}
	}

	return true
}
func validJSON(reading models.Event) bool {
	var js interface{}
	for i := range reading.Readings {
		err := json.Unmarshal([]byte(reading.Readings[i].Value), &js)
		if err != nil {
			loggingClient.Error("Error: ", err.Error())
			return false
		}
	}
	return true
}
