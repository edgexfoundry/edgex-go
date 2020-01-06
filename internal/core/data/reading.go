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
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/metadata"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

func getAllReadings(lc logger.LoggingClient, dbClient interfaces.DBClient,
	configuration *config.ConfigurationStruct) (readings []contract.Reading, err error) {

	readings, err = dbClient.Readings()
	if err != nil {
		lc.Error(err.Error())
		return nil, err
	}

	// Check max limit
	err = checkMaxLimit(len(readings), lc, configuration)
	if err != nil {
		return nil, err
	}

	return readings, nil
}

func decodeReading(
	reader io.Reader,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	configuration *config.ConfigurationStruct) (reading contract.Reading, err error) {

	reading = contract.Reading{}
	err = json.NewDecoder(reader).Decode(&reading)

	// Problem decoding
	if err != nil {
		lc.Error("Error decoding the reading: " + err.Error())

		return contract.Reading{}, errors.NewErrJsonDecoding(reading.Name)
	}

	if configuration.Writable.ValidateCheck {
		err = validateReading(reading, lc, dbClient)

		if err != nil {
			return contract.Reading{}, err
		}
	}

	return reading, nil
}

func validateReading(reading contract.Reading, lc logger.LoggingClient, dbClient interfaces.DBClient) error {
	// Check the value descriptor
	vd, err := dbClient.ValueDescriptorByName(reading.Name)
	if err != nil {
		lc.Error(err.Error())
		if err == db.ErrNotFound {
			return errors.NewErrDbNotFound()
		} else {
			return err
		}
	}

	err = isValidValueDescriptor(vd, reading)
	if err != nil {
		lc.Error(err.Error())
		return err
	}

	return nil
}

func addReading(
	reading contract.Reading,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) (id string, err error) {

	id, err = dbClient.AddReading(reading)

	if err != nil {
		lc.Error(err.Error())

		return "", err
	}

	return id, nil
}

func getReadingById(
	id string,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) (reading contract.Reading, err error) {

	reading, err = dbClient.ReadingById(id)

	if err != nil {
		lc.Error(err.Error())
		if err == db.ErrNotFound {
			return contract.Reading{}, errors.NewErrDbNotFound()
		} else {
			return contract.Reading{}, err
		}
	}

	return reading, nil
}

func deleteReadingById(id string, lc logger.LoggingClient, dbClient interfaces.DBClient) error {
	err := dbClient.DeleteReadingById(id)
	if err != nil {
		if err == db.ErrNotFound {
			return errors.NewErrDbNotFound()
		}
		lc.Error(err.Error())
		return err
	}

	return nil
}

func updateReading(
	reading contract.Reading,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	configuration *config.ConfigurationStruct) error {
	to, err := getReadingById(reading.Id, lc, dbClient)
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
		if configuration.Writable.ValidateCheck {
			fmt.Println(to)

			err = validateReading(to, lc, dbClient)
			if err != nil {
				lc.Error("Error validating updated reading")
				return err
			}
		}
	}

	err = dbClient.UpdateReading(reading)
	if err != nil {
		lc.Error(err.Error())
		return err
	}

	return nil
}

func countReadings(lc logger.LoggingClient, dbClient interfaces.DBClient) (count int, err error) {
	count, err = dbClient.ReadingCount()
	if err != nil {
		lc.Error(err.Error())
		return 0, err
	}

	return count, nil
}

func getReadingsByDevice(
	deviceId string,
	limit int,
	ctx context.Context,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	mdc metadata.DeviceClient,
	configuration *config.ConfigurationStruct) (readings []contract.Reading, err error) {

	if checkDevice(deviceId, ctx, mdc, configuration) != nil {
		lc.Error(fmt.Sprintf("error checking device %s %v", deviceId, err))

		return []contract.Reading{}, err
	}

	readings, err = dbClient.ReadingsByDevice(deviceId, limit)
	if err != nil {
		lc.Error(err.Error())
		return []contract.Reading{}, err
	}

	return readings, nil
}

func getReadingsByValueDescriptor(
	name string,
	limit int,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	configuration *config.ConfigurationStruct) (readings []contract.Reading, err error) {

	// Limit is too large
	err = checkMaxLimit(limit, lc, configuration)
	if err != nil {
		return []contract.Reading{}, err
	}

	// Check for value descriptor
	if configuration.Writable.ValidateCheck {
		_, err = getValueDescriptorByName(name, lc, dbClient)
		if err != nil {
			return []contract.Reading{}, err
		}
	}

	readings, err = dbClient.ReadingsByValueDescriptor(name, limit)
	if err != nil {
		lc.Error(err.Error())
		return []contract.Reading{}, err
	}

	return readings, nil
}

func getReadingsByValueDescriptorNames(
	listOfNames []string,
	limit int,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) (readings []contract.Reading, err error) {

	readings, err = dbClient.ReadingsByValueDescriptorNames(listOfNames, limit)
	if err != nil {
		lc.Error(err.Error())
		return nil, err
	}

	return readings, nil
}

func getReadingsByCreationTime(
	start int64,
	end int64,
	limit int,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) (readings []contract.Reading, err error) {

	readings, err = dbClient.ReadingsByCreationTime(start, end, limit)
	if err != nil {
		lc.Error(err.Error())
		return nil, err
	}

	return readings, nil
}

func getReadingsByDeviceAndValueDescriptor(
	device string,
	name string,
	limit int,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) (readings []contract.Reading, err error) {

	readings, err = dbClient.ReadingsByDeviceAndValueDescriptor(device, name, limit)
	if err != nil {
		lc.Error(err.Error())
		return nil, err
	}

	return readings, nil
}
