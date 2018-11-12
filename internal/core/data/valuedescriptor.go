/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"io"
	"regexp"

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

const (
	formatSpecifier          = "%(\\d+\\$)?([-#+ 0,(\\<]*)?(\\d+)?(\\.\\d+)?([tT])?([a-zA-Z%])"
	maxExceededString string = "Error, exceeded the max limit as defined in config"
)

// Check if the value descriptor matches the format string regular expression
func validateFormatString(v models.ValueDescriptor) (bool, error) {
	// No formatting specified
	if v.Formatting == "" {
		return true, nil
	} else {
		return regexp.MatchString(formatSpecifier, v.Formatting)
	}
}

func getValueDescriptorByName(name string) (vd models.ValueDescriptor, err error) {
	vd, err = dbClient.ValueDescriptorByName(name)

	if err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			return models.ValueDescriptor{}, errors.NewErrDbNotFound()
		} else {
			return models.ValueDescriptor{}, err
		}
	}

	return vd, nil
}

func getValueDescriptorById(id string) (vd models.ValueDescriptor, err error) {
	vd, err = dbClient.ValueDescriptorById(id)

	if err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			return models.ValueDescriptor{}, errors.NewErrDbNotFound()
		} else {
			return models.ValueDescriptor{}, err
		}
	}

	return vd, nil
}

func getValueDescriptorsByUomLabel(uomLabel string) (vdList []models.ValueDescriptor, err error) {
	vdList, err = dbClient.ValueDescriptorsByUomLabel(uomLabel)

	if err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			return []models.ValueDescriptor{}, errors.NewErrDbNotFound()
		} else {
			return []models.ValueDescriptor{}, err
		}
	}

	return vdList, nil
}

func getValueDescriptorsByLabel(label string) (vdList []models.ValueDescriptor, err error) {
	vdList, err = dbClient.ValueDescriptorsByLabel(label)

	if err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			return []models.ValueDescriptor{}, errors.NewErrDbNotFound()
		} else {
			return []models.ValueDescriptor{}, err
		}
	}

	return vdList, nil
}

func getValueDescriptorsByType(typ string)(vdList []models.ValueDescriptor, err error) {
	vdList, err = dbClient.ValueDescriptorsByType(typ)

	if err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			return []models.ValueDescriptor{}, errors.NewErrDbNotFound()
		} else {
			return []models.ValueDescriptor{}, err
		}
	}

	return vdList, nil
}

func getValueDescriptorsByDevice(device models.Device)(vdList []models.ValueDescriptor, err error) {
	// Get the names of the value descriptors
	vdNames := []string{}
	device.AllAssociatedValueDescriptors(&vdNames)

	// Get the value descriptors
	vdList = []models.ValueDescriptor{}
	for _, name := range vdNames {
		vd, err := getValueDescriptorByName(name)

		// Not an error if not found
		if err != nil {
			switch err.(type) {
			case *errors.ErrDbNotFound:
				continue
			default:
				return []models.ValueDescriptor{}, err
			}
		}

		vdList = append(vdList, vd)
	}

	return vdList, nil
}

func getValueDescriptorsByDeviceName(name string)(vdList []models.ValueDescriptor, err error) {
	// Get the device
	device, err := mdc.DeviceForName(name)
	if err != nil {
		switch err := err.(type) {
		case types.ErrNotFound:
			LoggingClient.Error("Device not found: " + err.Error())
			return []models.ValueDescriptor{}, errors.NewErrDbNotFound()
		default:
			LoggingClient.Error("Problem getting device from metadata: " + err.Error())
			return []models.ValueDescriptor{}, err
		}
	}

	return getValueDescriptorsByDevice(device)
}

func getValueDescriptorsByDeviceId(id string)(vdList []models.ValueDescriptor, err error) {
	// Get the device
	device, err := mdc.Device(id)
	if err != nil {
		switch err := err.(type) {
		case types.ErrNotFound:
			LoggingClient.Error("Device not found: " + err.Error())
			return []models.ValueDescriptor{}, errors.NewErrDbNotFound()
		default:
			LoggingClient.Error("Problem getting device from metadata: " + err.Error())
			return []models.ValueDescriptor{}, err
		}
	}

	return getValueDescriptorsByDevice(device)
}

func getAllValueDescriptors()(vd []models.ValueDescriptor, err error) {
	vd, err = dbClient.ValueDescriptors()
	if err != nil {
		LoggingClient.Error(err.Error())
		return nil, err
	}

	return vd, nil
}

func decodeValueDescriptor(reader io.ReadCloser) (vd models.ValueDescriptor, err error) {
	v := models.ValueDescriptor{}
	err = json.NewDecoder(reader).Decode(&v)
	// Problems decoding
	if err != nil {
		LoggingClient.Error("Error decoding the value descriptor: " + err.Error())
		return models.ValueDescriptor{}, errors.NewErrJsonDecoding(v.Name)
	}

	// Check the formatting
	match, err := validateFormatString(v)
	if err != nil {
		LoggingClient.Error("Error checking for format string for value descriptor " + v.Name)
		return models.ValueDescriptor{}, err
	}
	if !match {
		LoggingClient.Error("Error posting value descriptor. Format is not a valid printf format.")
		return models.ValueDescriptor{}, errors.NewErrValueDescriptorInvalid(v.Name, err)
	}

	return v, nil
}

func addValueDescriptor(vd models.ValueDescriptor) (id bson.ObjectId, err error) {
	id, err = dbClient.AddValueDescriptor(vd)
	if err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotUnique {
			return bson.ObjectId(""), errors.NewErrValueDescriptorInUse(vd.Name)
		} else {
			return bson.ObjectId(""), err
		}
	}

	return id, nil
}

func updateValueDescriptor(from models.ValueDescriptor) error {
	to, err := getValueDescriptorById(from.Id.Hex())
	if err != nil {
		return err
	}

	// Update the fields
	if from.DefaultValue != "" {
		to.DefaultValue = from.DefaultValue
	}
	if from.Formatting != "" {
		match, err := regexp.MatchString(formatSpecifier, from.Formatting)
		if err != nil {
			LoggingClient.Error("Error checking formatting for updated value descriptor")
			return err
		}
		if !match {
			LoggingClient.Error("value descriptor's format string doesn't fit the required pattern: " + formatSpecifier)
			return errors.NewErrValueDescriptorInvalid(from.Name, err)
		}
		to.Formatting = from.Formatting
	}
	if from.Labels != nil {
		to.Labels = from.Labels
	}

	if from.Max != "" {
		to.Max = from.Max
	}
	if from.Min != "" {
		to.Min = from.Min
	}
	if from.Name != "" {
		// Check if value descriptor is still in use by readings if the name changes
		if from.Name != to.Name {
			r, err := getReadingsByValueDescriptor(to.Name, 10) // Arbitrary limit, we're just checking if there are any readings
			if err != nil {
				LoggingClient.Error("Error checking the readings for the value descriptor: " + err.Error())
				return err
			}
			// Value descriptor is still in use
			if len(r) != 0 {
				LoggingClient.Error("Data integrity issue.  Value Descriptor with name:  " + from.Name + " is still referenced by existing readings.")
				return errors.NewErrValueDescriptorInUse(from.Name)
			}
		}
		to.Name = from.Name
	}
	if from.Origin != 0 {
		to.Origin = from.Origin
	}
	if from.Type != "" {
		to.Type = from.Type
	}
	if from.UomLabel != "" {
		to.UomLabel = from.UomLabel
	}

	// Push the updated valuedescriptor to the database
	err = dbClient.UpdateValueDescriptor(to)
	if err != nil {
		if err == db.ErrNotUnique {
			LoggingClient.Error("Value descriptor name is not unique")
			return errors.NewErrValueDescriptorInUse(to.Name)
		} else {
			LoggingClient.Error(err.Error())
			return err
		}
	}

	return nil
}

func deleteValueDescriptor(vd models.ValueDescriptor) error {
	// Check if the value descriptor is still in use by readings
	readings, err := dbClient.ReadingsByValueDescriptor(vd.Name, 10)
	if err != nil {
		LoggingClient.Error(err.Error())
		return err
	}
	if len(readings) > 0 {
		LoggingClient.Error("Data integrity issue.  Value Descriptor is still referenced by existing readings.")
		return errors.NewErrValueDescriptorInUse(vd.Name)
	}

	// Delete the value descriptor
	if err = dbClient.DeleteValueDescriptorById(vd.Id.Hex()); err != nil {
		LoggingClient.Error(err.Error())
		return err
	}

	return nil
}

func deleteValueDescriptorByName(name string) error {
	// Check if the value descriptor exists
	vd, err := getValueDescriptorByName(name)
	if err != nil {
		return err
	}

	if err = deleteValueDescriptor(vd); err != nil {
		return err
	}

	return nil
}

func deleteValueDescriptorById(id string) error {
	// Check if the value descriptor exists
	vd, err := getValueDescriptorById(id)
	if err != nil {
		return err
	}

	if err = deleteValueDescriptor(vd); err != nil {
		return err
	}

	return nil
}
