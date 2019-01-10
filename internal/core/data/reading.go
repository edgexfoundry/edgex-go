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
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"io"
)

func getAllReadings() (readings []contract.Reading, err error) {
	readings, err = dbClient.Readings()
	if err != nil {
		LoggingClient.Error(err.Error())
		return nil, err
	}

	// Check max limit
	err = checkMaxLimit(len(readings))
	if err != nil {
		return nil, err
	}

	return readings, nil
}

func decodeReading(reader io.Reader) (reading contract.Reading, err error) {
	reading = contract.Reading{}
	err = json.NewDecoder(reader).Decode(&reading)

	// Problem decoding
	if err != nil {
		LoggingClient.Error("Error decoding the reading: " + err.Error())

		return contract.Reading{}, errors.NewErrJsonDecoding(reading.Name)
	}

	if Configuration.Writable.ValidateCheck {
		err = validateReading(reading)

		if err != nil {
			return contract.Reading{}, err
		}
	}

	return reading, nil
}

func validateReading(reading contract.Reading) error {
	// Check the value descriptor
	vd, err := dbClient.ValueDescriptorByName(reading.Name)
	if err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			return errors.NewErrDbNotFound()
		} else {
			return err
		}
	}

	err = isValidValueDescriptor(vd, reading)
	if err != nil {
		LoggingClient.Error(err.Error())
		return err
	}

	return nil
}

func addReading(reading contract.Reading) (id string, err error) {
	id, err = dbClient.AddReading(reading)

	if err != nil {
		LoggingClient.Error(err.Error())

		return "", err
	}

	return id, nil
}

func getReadingById(id string) (reading contract.Reading, err error) {
	reading, err = dbClient.ReadingById(id)

	if err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			return contract.Reading{}, errors.NewErrDbNotFound()
		} else {
			return contract.Reading{}, err
		}
	}

	return reading, nil
}

func getReadingsByDeviceId(limit int, deviceId string, valueDescriptor string) ([]contract.Reading, error) {
	eventList, err := dbClient.EventsForDevice(deviceId)
	if err != nil {
		LoggingClient.Error(err.Error())
		return nil, err
	}

	// Only pick the readings who match the value descriptor
	var readings []contract.Reading
	count := 0 // Make sure we stay below the limit
	for _, event := range eventList {
		if count >= limit {
			break
		}
		for _, reading := range event.Readings {
			if count >= limit {
				break
			}
			if reading.Name == valueDescriptor {
				readings = append(readings, reading)
				count += 1
			}
		}
	}

	return readings, nil
}

func deleteReadingById(id string) error {
	err := dbClient.DeleteReadingById(id)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NewErrDbNotFound()
		}
		LoggingClient.Error(err.Error())
		return err
	}

	return nil
}

func updateReading(reading contract.Reading) error {
	to, err := getReadingById(reading.Id)
	if err != nil {
		return err
	}

	// Update the fields
	if reading.Value != "" {
		to.Value = reading.Value
	}
	if reading.Name != "" {
		to.Name = reading.Name
	}
	if reading.Origin != 0 {
		to.Origin = reading.Origin
	}

	if reading.Value != "" || reading.Name != "" {
		if Configuration.Writable.ValidateCheck {
			fmt.Println(to)

			err = validateReading(to)
			if err != nil {
				LoggingClient.Error("Error validating updated reading")
				return err
			}
		}
	}

	err = dbClient.UpdateReading(reading)
	if err != nil {
		LoggingClient.Error(err.Error())
		return err
	}

	return nil
}

func countReadings() (count int, err error) {
	count, err = dbClient.ReadingCount()
	if err != nil {
		LoggingClient.Error(err.Error())
		return 0, err
	}

	return count, nil
}

func getReadingsByDevice(deviceId string, limit int) (readings []contract.Reading, err error) {
	if checkDevice(deviceId) != nil {
		LoggingClient.Error(fmt.Sprintf("error checking device %s %v", deviceId, err))

		return []contract.Reading{}, err
	}

	readings, err = dbClient.ReadingsByDevice(deviceId, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return []contract.Reading{}, err
	}

	return readings, nil
}

func getReadingsByValueDescriptor(name string, limit int) (readings []contract.Reading, err error) {
	// Limit is too large
	err = checkMaxLimit(limit)
	if err != nil {
		return []contract.Reading{}, err
	}

	// Check for value descriptor
	if Configuration.Writable.ValidateCheck {
		_, err = getValueDescriptorByName(name)
		if err != nil {
			return []contract.Reading{}, err
		}
	}

	readings, err = dbClient.ReadingsByValueDescriptor(name, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return []contract.Reading{}, err
	}

	return readings, nil
}

func getReadingsByValueDescriptorNames(listOfNames []string, limit int) (readings []contract.Reading, err error) {
	readings, err = dbClient.ReadingsByValueDescriptorNames(listOfNames, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return nil, err
	}

	return readings, nil
}

func getReadingsByCreationTime(start int64, end int64, limit int) (readings []contract.Reading, err error) {
	readings, err = dbClient.ReadingsByCreationTime(start, end, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return nil, err
	}

	return readings, nil
}

func getReadingsByDeviceAndValueDescriptor(device string, name string, limit int) (readings []contract.Reading, err error) {
	readings, err = dbClient.ReadingsByDeviceAndValueDescriptor(device, name, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return nil, err
	}

	return readings, nil
}
