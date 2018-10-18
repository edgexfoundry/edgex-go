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

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	models "github.com/edgexfoundry/edgex-go/pkg/models"
)

func isValidValueDescriptor(vd models.ValueDescriptor, reading models.Reading) error {
	var err error
	switch vd.Type {
	case "B": // boolean
		err = validBoolean(reading)
	case "F": // floating point
		err = validFloat(reading, vd)
	case "I": // integer
		err = validInteger(reading, vd)
	case "S": // string or character data
		err = validString(reading)
	case "J": // JSON data
		err = validJSON(reading)
	default:
		err = fmt.Errorf("Unknown type")
	}
	if err != nil {
		return errors.NewErrValueDescriptorInvalid(vd.Name, err)
	}
	return nil
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
