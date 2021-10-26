//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	pkgCommon "github.com/edgexfoundry/edgex-go/internal/pkg/common"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

const (
	ReadingsCollection                       = "cd|rd"
	ReadingsCollectionOrigin                 = ReadingsCollection + DBKeySeparator + common.Origin
	ReadingsCollectionDeviceName             = ReadingsCollection + DBKeySeparator + common.DeviceName
	ReadingsCollectionResourceName           = ReadingsCollection + DBKeySeparator + common.ResourceName
	ReadingsCollectionDeviceNameResourceName = ReadingsCollection + DBKeySeparator + common.DeviceName + DBKeySeparator + common.ResourceName
)

var emptyBinaryValue = make([]byte, 0)

// asyncDeleteReadingsByIds deletes all readings with given reading Ids.  This function is implemented to be run as a
// separate gorountine in the background to achieve better performance, so this function return nothing.  When
// encountering any errors during deletion, this function will simply log the error.
func (c *Client) asyncDeleteReadingsByIds(readingIds []string) {
	conn := c.Pool.Get()
	defer conn.Close()

	var readings [][]byte
	//start a transaction to get all readings
	readings, edgeXerr := getObjectsByIds(conn, pkgCommon.ConvertStringsToInterfaces(readingIds))
	if edgeXerr != nil {
		c.loggingClient.Error(fmt.Sprintf("Deleted readings failed while retrieving objects by Ids.  Err: %s", edgeXerr.DebugMessages()))
		return
	}

	// iterate each readings for deletion in batch
	queriesInQueue := 0
	r := models.BaseReading{}
	_ = conn.Send(MULTI)
	for i, reading := range readings {
		err := json.Unmarshal(reading, &r)
		if err != nil {
			c.loggingClient.Error(fmt.Sprintf("unable to marshal reading.  Err: %s", err.Error()))
			continue
		}
		storedKey := readingStoredKey(r.Id)
		_ = conn.Send(UNLINK, storedKey)
		_ = conn.Send(ZREM, ReadingsCollection, storedKey)
		_ = conn.Send(ZREM, ReadingsCollectionOrigin, storedKey)
		_ = conn.Send(ZREM, CreateKey(ReadingsCollectionDeviceName, r.DeviceName), storedKey)
		_ = conn.Send(ZREM, CreateKey(ReadingsCollectionResourceName, r.ResourceName), storedKey)
		_ = conn.Send(ZREM, CreateKey(ReadingsCollectionDeviceNameResourceName, r.DeviceName, r.ResourceName), storedKey)
		queriesInQueue++

		if queriesInQueue >= c.BatchSize {
			_, err = conn.Do(EXEC)
			if err != nil {
				c.loggingClient.Error(fmt.Sprintf("unable to execute batch reading deletion.  Err: %s", err.Error()))
				continue
			}
			// reset queriesInQueue to zero if EXEC is successfully executed without error
			queriesInQueue = 0
			// rerun another transaction when reading iteration is not finished
			if i < len(readings)-1 {
				_ = conn.Send(MULTI)
			}
		}
	}

	if queriesInQueue > 0 {
		_, err := conn.Do(EXEC)
		if err != nil {
			c.loggingClient.Error(fmt.Sprintf("unable to execute batch reading deletion.  Err: %s", err.Error()))
		}
	}
}

// readingStoredKey return the reading's stored key which combines the collection name and object id
func readingStoredKey(id string) string {
	return CreateKey(ReadingsCollection, id)
}

// Add a reading to the database
func addReading(conn redis.Conn, r models.Reading) (reading models.Reading, edgeXerr errors.EdgeX) {
	var m []byte
	var err error
	var baseReading *models.BaseReading
	switch newReading := r.(type) {
	case models.BinaryReading:
		// Clear the binary data since we do not want to persist binary data to save on memory.
		newReading.BinaryValue = emptyBinaryValue

		baseReading = &newReading.BaseReading
		if err = checkReadingValue(baseReading); err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
		m, err = json.Marshal(newReading)
		reading = newReading
	case models.SimpleReading:
		baseReading = &newReading.BaseReading
		if err = checkReadingValue(baseReading); err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
		m, err = json.Marshal(newReading)
		reading = newReading
	case models.ObjectReading:
		baseReading = &newReading.BaseReading
		if err = checkReadingValue(baseReading); err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
		m, err = json.Marshal(newReading)
		reading = newReading
	default:
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "unsupported reading type", nil)
	}

	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "reading parsing failed", err)
	}
	storedKey := readingStoredKey(baseReading.Id)
	// use the SET command to save reading as blob
	_ = conn.Send(SET, storedKey, m)
	_ = conn.Send(ZADD, ReadingsCollection, 0, storedKey)
	_ = conn.Send(ZADD, ReadingsCollectionOrigin, baseReading.Origin, storedKey)
	_ = conn.Send(ZADD, CreateKey(ReadingsCollectionDeviceName, baseReading.DeviceName), baseReading.Origin, storedKey)
	_ = conn.Send(ZADD, CreateKey(ReadingsCollectionResourceName, baseReading.ResourceName), baseReading.Origin, storedKey)
	_ = conn.Send(ZADD, CreateKey(ReadingsCollectionDeviceNameResourceName, baseReading.DeviceName, baseReading.ResourceName), baseReading.Origin, storedKey)

	return reading, nil
}

// Remove a reading out of the database
func deleteReadingById(conn redis.Conn, id string) (edgeXerr errors.EdgeX) {
	r := models.BaseReading{}
	storedKey := readingStoredKey(id)
	edgeXerr = getObjectById(conn, storedKey, &r)
	if edgeXerr != nil {
		return edgeXerr
	}

	_ = conn.Send(MULTI)
	_ = conn.Send(UNLINK, storedKey)
	_ = conn.Send(ZREM, ReadingsCollection, storedKey)
	_ = conn.Send(ZREM, ReadingsCollectionOrigin, storedKey)
	_ = conn.Send(ZREM, CreateKey(ReadingsCollectionDeviceName, r.DeviceName), storedKey)
	_ = conn.Send(ZREM, CreateKey(ReadingsCollectionResourceName, r.ResourceName), storedKey)
	_ = conn.Send(ZREM, CreateKey(ReadingsCollectionDeviceNameResourceName, r.DeviceName, r.ResourceName), storedKey)
	_, err := conn.Do(EXEC)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("reading[id:%s] delete failed", id), err)
	}

	return nil
}

func checkReadingValue(b *models.BaseReading) errors.EdgeX {
	// check if id is a valid uuid
	if b.Id == "" {
		b.Id = uuid.New().String()
	} else {
		_, err := uuid.Parse(b.Id)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindInvalidId, "uuid parsing failed", err)
		}
	}
	return nil
}

func readingsByEventId(conn redis.Conn, eventId string) (readings []models.Reading, edgeXerr errors.EdgeX) {
	objects, err := getObjectsByRange(conn, CreateKey(EventsCollectionReadings, eventId), 0, -1)
	if errors.Kind(err) == errors.KindEntityDoesNotExist {
		return // Empty Readings in an Event is not an error
	} else if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}

	return convertObjectsToReadings(objects)
}

func allReadings(conn redis.Conn, offset int, limit int) (readings []models.Reading, edgeXerr errors.EdgeX) {
	objects, err := getObjectsByRevRange(conn, ReadingsCollectionOrigin, offset, limit)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}

	return convertObjectsToReadings(objects)
}

// readingsByResourceName query readings by offset, limit, and resource name
func readingsByResourceName(conn redis.Conn, offset int, limit int, resourceName string) (readings []models.Reading, edgeXerr errors.EdgeX) {
	objects, err := getObjectsByRevRange(conn, CreateKey(ReadingsCollectionResourceName, resourceName), offset, limit)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}

	return convertObjectsToReadings(objects)
}

// readingsByDeviceName query readings by offset, limit, and device name
func readingsByDeviceName(conn redis.Conn, offset int, limit int, name string) (readings []models.Reading, edgeXerr errors.EdgeX) {
	objects, err := getObjectsByRevRange(conn, CreateKey(ReadingsCollectionDeviceName, name), offset, limit)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}

	return convertObjectsToReadings(objects)
}

func readingsByDeviceNameAndResourceName(conn redis.Conn, deviceName string, resourceName string, offset int, limit int) (readings []models.Reading, err errors.EdgeX) {
	objects, err := getObjectsByRevRange(conn, CreateKey(ReadingsCollectionDeviceNameResourceName, deviceName, resourceName), offset, limit)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}

	return convertObjectsToReadings(objects)
}

func readingsByDeviceNameAndResourceNameAndTimeRange(conn redis.Conn, deviceName string, resourceName string, startTime int, endTime int, offset int, limit int) (readings []models.Reading, err errors.EdgeX) {
	objects, err := getObjectsByScoreRange(conn, CreateKey(ReadingsCollectionDeviceNameResourceName, deviceName, resourceName), startTime, endTime, offset, limit)
	if err != nil {
		return readings, err
	}

	return convertObjectsToReadings(objects)
}

func readingsByDeviceNameAndResourceNamesAndTimeRange(conn redis.Conn, deviceName string, resourceNames []string, startTime int, endTime int, offset int, limit int) (readings []models.Reading, totalCount uint32, err errors.EdgeX) {
	var redisKeys []string
	for _, resourceName := range resourceNames {
		redisKeys = append(redisKeys, CreateKey(ReadingsCollectionDeviceNameResourceName, deviceName, resourceName))
	}

	objects, totalCount, err := unionObjectsByKeysAndScoreRange(conn, startTime, endTime, offset, limit, redisKeys...)
	if err != nil {
		return readings, totalCount, err
	}
	readings, err = convertObjectsToReadings(objects)
	return readings, totalCount, err
}

// readingsByTimeRange query readings by time range, offset, and limit
func readingsByTimeRange(conn redis.Conn, startTime int, endTime int, offset int, limit int) (readings []models.Reading, edgeXerr errors.EdgeX) {
	objects, edgeXerr := getObjectsByScoreRange(conn, ReadingsCollectionOrigin, startTime, endTime, offset, limit)
	if edgeXerr != nil {
		return readings, edgeXerr
	}
	return convertObjectsToReadings(objects)
}

func readingsByResourceNameAndTimeRange(conn redis.Conn, resourceName string, startTime int, endTime int, offset int, limit int) (readings []models.Reading, edgeXerr errors.EdgeX) {
	objects, edgeXerr := getObjectsByScoreRange(conn, CreateKey(ReadingsCollectionResourceName, resourceName), startTime, endTime, offset, limit)
	if edgeXerr != nil {
		return readings, edgeXerr
	}
	return convertObjectsToReadings(objects)
}

func readingsByDeviceNameAndTimeRange(conn redis.Conn, deviceName string, startTime int, endTime int, offset int, limit int) (readings []models.Reading, edgeXerr errors.EdgeX) {
	objects, edgeXerr := getObjectsByScoreRange(conn, CreateKey(ReadingsCollectionDeviceName, deviceName), startTime, endTime, offset, limit)
	if edgeXerr != nil {
		return readings, edgeXerr
	}
	return convertObjectsToReadings(objects)
}

func convertObjectsToReadings(objects [][]byte) (readings []models.Reading, edgeXerr errors.EdgeX) {
	readings = make([]models.Reading, len(objects))
	var alias struct {
		ValueType string
	}
	for i, in := range objects {
		err := json.Unmarshal(in, &alias)
		if err != nil {
			return []models.Reading{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "reading format parsing failed from the database", err)
		}
		if alias.ValueType == common.ValueTypeBinary {
			var binaryReading models.BinaryReading
			err = json.Unmarshal(in, &binaryReading)
			if err != nil {
				return []models.Reading{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "binary reading format parsing failed from the database", err)
			}
			readings[i] = binaryReading
		} else if alias.ValueType == common.ValueTypeObject {
			var objectReading models.ObjectReading
			err = json.Unmarshal(in, &objectReading)
			if err != nil {
				return []models.Reading{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "object reading format parsing failed from the database", err)
			}
			readings[i] = objectReading
		} else {
			var simpleReading models.SimpleReading
			err = json.Unmarshal(in, &simpleReading)
			if err != nil {
				return []models.Reading{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "simple reading format parsing failed from the database", err)
			}
			readings[i] = simpleReading
		}
	}
	return readings, nil
}
