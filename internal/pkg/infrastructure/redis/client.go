//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	redisClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/redis"

	"github.com/google/uuid"
)

type Client struct {
	*redisClient.Client
	loggingClient logger.LoggingClient
}

func NewClient(config db.Configuration, logger logger.LoggingClient) (*Client, errors.EdgeX) {
	var err error
	dc := &Client{}
	dc.Client, err = redisClient.NewClient(config, logger)
	dc.loggingClient = logger
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "redis client creation failed", err)
	}

	return dc, nil
}

func closeRedisConnection(conn redis.Conn, logger logger.LoggingClient) {
	err := conn.Close()
	if err != nil {
		logger.Errorf("error occured while closing the redis connection: %s", err.Error())
	}
}

// AddEvent adds a new event
func (c *Client) AddEvent(e model.Event) (model.Event, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)
	if e.Id != "" {
		_, err := uuid.Parse(e.Id)
		if err != nil {
			return model.Event{}, errors.NewCommonEdgeX(errors.KindInvalidId, "uuid parsing failed", err)
		}
	}

	return addEvent(conn, e)
}

// EventById gets an event by id
func (c *Client) EventById(id string) (event model.Event, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	event, edgeXerr = eventById(conn, id)
	if edgeXerr != nil {
		return event, errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to query event by id %s", id), edgeXerr)
	}

	return
}

// DeleteEventById removes an event by id
func (c *Client) DeleteEventById(id string) (edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	edgeXerr = deleteEventById(conn, id)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return
}

// Add a new device profle
func (c *Client) AddDeviceProfile(dp model.DeviceProfile) (model.DeviceProfile, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	if dp.Id != "" {
		_, err := uuid.Parse(dp.Id)
		if err != nil {
			return model.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindInvalidId, "ID failed UUID parsing", err)
		}
	} else {
		dp.Id = uuid.New().String()
	}

	return addDeviceProfile(conn, dp)
}

// UpdateDeviceProfile updates a new device profile
func (c *Client) UpdateDeviceProfile(dp model.DeviceProfile) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)
	return updateDeviceProfile(conn, dp)
}

// DeviceProfileNameExists checks the device profile exists by name
func (c *Client) DeviceProfileNameExists(name string) (bool, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)
	return deviceProfileNameExists(conn, name)
}

// AddDeviceService adds a new device service
func (c *Client) AddDeviceService(ds model.DeviceService) (model.DeviceService, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	if len(ds.Id) == 0 {
		ds.Id = uuid.New().String()
	}

	return addDeviceService(conn, ds)
}

// DeviceServiceByName gets a device service by name
func (c *Client) DeviceServiceByName(name string) (deviceService model.DeviceService, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	deviceService, edgeXerr = deviceServiceByName(conn, name)
	if edgeXerr != nil {
		return deviceService, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return
}

// DeviceServiceById gets a device service by id
func (c *Client) DeviceServiceById(id string) (deviceService model.DeviceService, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	deviceService, edgeXerr = deviceServiceById(conn, id)
	if edgeXerr != nil {
		return deviceService, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return
}

// DeleteDeviceServiceById deletes a device service by id
func (c *Client) DeleteDeviceServiceById(id string) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	edgeXerr := deleteDeviceServiceById(conn, id)
	if edgeXerr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to delete the device service with id %s", id), edgeXerr)
	}

	return nil
}

// DeleteDeviceServiceByName deletes a device service by name
func (c *Client) DeleteDeviceServiceByName(name string) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	edgeXerr := deleteDeviceServiceByName(conn, name)
	if edgeXerr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to delete the device service with name %s", name), edgeXerr)
	}

	return nil
}

// DeviceServiceNameExists checks the device service exists by name
func (c *Client) DeviceServiceNameExists(name string) (bool, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)
	return deviceServiceNameExist(conn, name)
}

// UpdateDeviceService updates a device service
func (c *Client) UpdateDeviceService(ds model.DeviceService) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)
	return updateDeviceService(conn, ds)
}

// DeviceProfileById gets a device profile by id
func (c *Client) DeviceProfileById(id string) (deviceProfile model.DeviceProfile, err errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	deviceProfile, err = deviceProfileById(conn, id)
	if err != nil {
		return deviceProfile, errors.NewCommonEdgeXWrapper(err)
	}

	return
}

// DeviceProfileByName gets a device profile by name
func (c *Client) DeviceProfileByName(name string) (deviceProfile model.DeviceProfile, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	deviceProfile, edgeXerr = deviceProfileByName(conn, name)
	if edgeXerr != nil {
		return deviceProfile, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return
}

// DeleteDeviceProfileById deletes a device profile by id
func (c *Client) DeleteDeviceProfileById(id string) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	edgeXerr := deleteDeviceProfileById(conn, id)
	if edgeXerr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to delete the device profile with id %s", id), edgeXerr)
	}

	return nil
}

// DeleteDeviceProfileByName deletes a device profile by name
func (c *Client) DeleteDeviceProfileByName(name string) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	edgeXerr := deleteDeviceProfileByName(conn, name)
	if edgeXerr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to delete the device profile with name %s", name), edgeXerr)
	}

	return nil
}

// AllDeviceProfiles query device profiles with offset and limit
func (c *Client) AllDeviceProfiles(offset int, limit int, labels []string) ([]model.DeviceProfile, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	deviceProfiles, edgeXerr := deviceProfilesByLabels(conn, offset, limit, labels)
	if edgeXerr != nil {
		return deviceProfiles, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return deviceProfiles, nil
}

// DeviceProfilesByModel query device profiles with offset, limit and model
func (c *Client) DeviceProfilesByModel(offset int, limit int, model string) ([]model.DeviceProfile, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	deviceProfiles, edgeXerr := deviceProfilesByModel(conn, offset, limit, model)
	if edgeXerr != nil {
		return deviceProfiles, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return deviceProfiles, nil
}

// DeviceProfilesByManufacturer query device profiles with offset, limit and manufacturer
func (c *Client) DeviceProfilesByManufacturer(offset int, limit int, manufacturer string) ([]model.DeviceProfile, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	deviceProfiles, edgeXerr := deviceProfilesByManufacturer(conn, offset, limit, manufacturer)
	if edgeXerr != nil {
		return deviceProfiles, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return deviceProfiles, nil
}

// DeviceProfilesByManufacturerAndModel query device profiles with offset, limit, manufacturer and model
func (c *Client) DeviceProfilesByManufacturerAndModel(offset int, limit int, manufacturer string, model string) ([]model.DeviceProfile, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	deviceProfiles, edgeXerr := deviceProfilesByManufacturerAndModel(conn, offset, limit, manufacturer, model)
	if edgeXerr != nil {
		return deviceProfiles, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return deviceProfiles, nil
}

// EventTotalCount returns the total count of Event from the database
func (c *Client) EventTotalCount() (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, EventsCollection)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// EventCountByDeviceName returns the count of Event associated a specific Device from the database
func (c *Client) EventCountByDeviceName(deviceName string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(EventsCollectionDeviceName, deviceName))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// EventCountByTimeRange returns the count of Event by time range
func (c *Client) EventCountByTimeRange(startTime int64, endTime int64) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberCountByScoreRange(conn, EventsCollectionOrigin, startTime, endTime)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// AllDeviceServices returns multiple device services per query criteria, including
// offset: the number of items to skip before starting to collect the result set
// limit: The numbers of items to return
// labels: allows for querying a given object by associated user-defined labels
func (c *Client) AllDeviceServices(offset int, limit int, labels []string) (deviceServices []model.DeviceService, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	deviceServices, edgeXerr = deviceServicesByLabels(conn, offset, limit, labels)
	if edgeXerr != nil {
		return deviceServices, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return deviceServices, nil
}

// Add a new device
func (c *Client) AddDevice(d model.Device) (model.Device, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	if len(d.Id) == 0 {
		d.Id = uuid.New().String()
	}

	return addDevice(conn, d)
}

// DeleteDeviceById deletes a device by id
func (c *Client) DeleteDeviceById(id string) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	edgeXerr := deleteDeviceById(conn, id)
	if edgeXerr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to delete the device with id %s", id), edgeXerr)
	}

	return nil
}

// DeleteDeviceByName deletes a device by name
func (c *Client) DeleteDeviceByName(name string) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	edgeXerr := deleteDeviceByName(conn, name)
	if edgeXerr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to delete the device with name %s", name), edgeXerr)
	}

	return nil
}

// DevicesByServiceName query devices by offset, limit and name
func (c *Client) DevicesByServiceName(offset int, limit int, name string) (devices []model.Device, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	devices, edgeXerr = devicesByServiceName(conn, offset, limit, name)
	if edgeXerr != nil {
		return devices, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query devices by offset %d, limit %d and name %s", offset, limit, name), edgeXerr)
	}
	return devices, nil
}

// DeviceIdExists checks the device existence by id
func (c *Client) DeviceIdExists(id string) (bool, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)
	exists, err := deviceIdExists(conn, id)
	if err != nil {
		return exists, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("fail to check the device existence by id %s", id), err)
	}
	return exists, nil
}

// DeviceNameExists checks the device existence by name
func (c *Client) DeviceNameExists(name string) (bool, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)
	exists, err := deviceNameExists(conn, name)
	if err != nil {
		return exists, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("fail to check the device existence by name %s", name), err)
	}
	return exists, nil
}

// DeviceById gets a device by id
func (c *Client) DeviceById(id string) (device model.Device, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	device, edgeXerr = deviceById(conn, id)
	if edgeXerr != nil {
		return device, errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to query device by id %s", id), edgeXerr)
	}

	return
}

// DeviceByName gets a device by name
func (c *Client) DeviceByName(name string) (device model.Device, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	device, edgeXerr = deviceByName(conn, name)
	if edgeXerr != nil {
		return device, errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to query device by name %s", name), edgeXerr)
	}

	return
}

// DevicesByProfileName query devices by offset, limit and profile name
func (c *Client) DevicesByProfileName(offset int, limit int, profileName string) (devices []model.Device, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	devices, edgeXerr = devicesByProfileName(conn, offset, limit, profileName)
	if edgeXerr != nil {
		return devices, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query devices by offset %d, limit %d and name %s", offset, limit, profileName), edgeXerr)
	}
	return devices, nil
}

// Update a device
func (c *Client) UpdateDevice(d model.Device) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	return updateDevice(conn, d)
}

// AllEvents query events by offset and limit
func (c *Client) AllEvents(offset int, limit int) ([]model.Event, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	events, edgeXerr := c.allEvents(conn, offset, limit)
	if edgeXerr != nil {
		return events, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query events by offset %d and limit %d", offset, limit), edgeXerr)
	}
	return events, nil
}

// AllDevices query the devices with offset, limit, and labels
func (c *Client) AllDevices(offset int, limit int, labels []string) ([]model.Device, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	devices, edgeXerr := devicesByLabels(conn, offset, limit, labels)
	if edgeXerr != nil {
		return devices, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return devices, nil
}

func (c *Client) DeviceTree(parent string, levels int, offset int, limit int, labels []string) (uint32, []model.Device, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	totalCount, devices, edgeXerr := deviceTree(conn, parent, levels, offset, limit, labels)
	if edgeXerr != nil {
		return totalCount, devices, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return totalCount, devices, nil
}

// EventsByDeviceName query events by offset, limit and device name
func (c *Client) EventsByDeviceName(offset int, limit int, name string) (events []model.Event, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	events, edgeXerr = eventsByDeviceName(conn, offset, limit, name)
	if edgeXerr != nil {
		return events, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query events by offset %d, limit %d and name %s", offset, limit, name), edgeXerr)
	}
	return events, nil
}

// EventsByTimeRange query events by time range, offset, and limit
func (c *Client) EventsByTimeRange(startTime int64, endTime int64, offset int, limit int) (events []model.Event, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	events, edgeXerr = eventsByTimeRange(conn, startTime, endTime, offset, limit)
	if edgeXerr != nil {
		return events, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query events by time range %v ~ %v, offset %d, and limit %d", startTime, endTime, offset, limit), edgeXerr)
	}
	return events, nil
}

// ReadingTotalCount returns the total count of Event from the database
func (c *Client) ReadingTotalCount() (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, ReadingsCollection)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// AllReadings query events by offset, limit, and labels
func (c *Client) AllReadings(offset int, limit int) ([]model.Reading, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	readings, edgeXerr := allReadings(conn, offset, limit)
	if edgeXerr != nil {
		return readings, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query readings by offset %d, and limit %d", offset, limit), edgeXerr)
	}
	return readings, nil
}

// ReadingsByTimeRange query readings by time range, offset, and limit
func (c *Client) ReadingsByTimeRange(start int64, end int64, offset int, limit int) (readings []model.Reading, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	readings, edgeXerr = readingsByTimeRange(conn, start, end, offset, limit)
	if edgeXerr != nil {
		return readings, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query readings by time range %v ~ %v, offset %d, and limit %d", start, end, offset, limit), edgeXerr)
	}
	return readings, nil
}

// ReadingsByResourceName query readings by offset, limit and resource name
func (c *Client) ReadingsByResourceName(offset int, limit int, resourceName string) (readings []model.Reading, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	readings, edgeXerr = readingsByResourceName(conn, offset, limit, resourceName)
	if edgeXerr != nil {
		return readings, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query readings by offset %d, limit %d and resourceName %s", offset, limit, resourceName), edgeXerr)
	}
	return readings, nil
}

// ReadingsByDeviceName query readings by offset, limit and device name
func (c *Client) ReadingsByDeviceName(offset int, limit int, name string) (readings []model.Reading, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	readings, edgeXerr = readingsByDeviceName(conn, offset, limit, name)
	if edgeXerr != nil {
		return readings, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query readings by offset %d, limit %d and name %s", offset, limit, name), edgeXerr)
	}
	return readings, nil
}

// ReadingCountByDeviceName returns the count of Readings associated a specific Device from the database
func (c *Client) ReadingCountByDeviceName(deviceName string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(ReadingsCollectionDeviceName, deviceName))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// ReadingCountByResourceName returns the count of Readings associated a specific Resource from the database
func (c *Client) ReadingCountByResourceName(resourceName string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(ReadingsCollectionResourceName, resourceName))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// ReadingCountByResourceNameAndTimeRange returns the count of Readings associated a specific Resource from the database within specified time range
func (c *Client) ReadingCountByResourceNameAndTimeRange(resourceName string, startTime int64, endTime int64) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberCountByScoreRange(conn, CreateKey(ReadingsCollectionResourceName, resourceName), startTime, endTime)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// ReadingCountByDeviceNameAndResourceName returns the count of Readings associated with specified Resource and Device from the database
func (c *Client) ReadingCountByDeviceNameAndResourceName(deviceName string, resourceName string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(ReadingsCollectionDeviceNameResourceName, deviceName, resourceName))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// ReadingCountByDeviceNameAndResourceNameAndTimeRange returns the count of Readings associated with specified Resource and Device from the database within specified time range
func (c *Client) ReadingCountByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, start int64, end int64) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberCountByScoreRange(conn, CreateKey(ReadingsCollectionDeviceNameResourceName, deviceName, resourceName), start, end)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// ReadingCountByDeviceNameAndResourceNamesAndTimeRange returns the count of readings by origin within the time range
// associated with the specified device and resourceName slice from db
func (c *Client) ReadingCountByDeviceNameAndResourceNamesAndTimeRange(deviceName string, resourceNames []string, start int64, end int64) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	readings, err := readingsByDeviceNameAndResourceNamesAndTimeRange(conn, deviceName, resourceNames, start, end, 0, -1)
	if err != nil {
		return 0, errors.NewCommonEdgeX(errors.Kind(err),
			fmt.Sprintf("fail to query readings by deviceName %s, resourceNames %v and time range %v ~ %v", deviceName, resourceNames, start, end), err)
	}

	return uint32(len(readings)), nil
}

// ReadingCountByTimeRange returns the count of Readings from the database within specified time range
func (c *Client) ReadingCountByTimeRange(start int64, end int64) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberCountByScoreRange(conn, ReadingsCollectionOrigin, start, end)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// ReadingsByResourceNameAndTimeRange query readings by resourceName and specified time range. Readings are sorted in descending order of origin time.
func (c *Client) ReadingsByResourceNameAndTimeRange(resourceName string, start int64, end int64, offset int, limit int) (readings []model.Reading, err errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	readings, err = readingsByResourceNameAndTimeRange(conn, resourceName, start, end, offset, limit)
	if err != nil {
		return readings, errors.NewCommonEdgeX(errors.Kind(err),
			fmt.Sprintf("fail to query readings by resourceName %s and time range %v ~ %v, offset %d, and limit %d", resourceName, start, end, offset, limit), err)
	}
	return readings, nil
}

func (c *Client) ReadingsByDeviceNameAndResourceName(deviceName string, resourceName string, offset int, limit int) (readings []model.Reading, err errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	readings, err = readingsByDeviceNameAndResourceName(conn, deviceName, resourceName, offset, limit)
	if err != nil {
		return readings, errors.NewCommonEdgeX(errors.Kind(err),
			fmt.Sprintf("fail to query readings by deviceName %s and resourceName %s", deviceName, resourceName), err)
	}

	return readings, nil
}

func (c *Client) ReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, start int64, end int64, offset int, limit int) (readings []model.Reading, err errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	readings, err = readingsByDeviceNameAndResourceNameAndTimeRange(conn, deviceName, resourceName, start, end, offset, limit)
	if err != nil {
		return readings, errors.NewCommonEdgeX(errors.Kind(err),
			fmt.Sprintf("fail to query readings by deviceName %s, resourceName %s and time range %v ~ %v", deviceName, resourceName, start, end), err)
	}

	return readings, nil
}

func (c *Client) ReadingsByDeviceNameAndResourceNamesAndTimeRange(deviceName string, resourceNames []string, start, end int64, offset, limit int) (readings []model.Reading, err errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	readings, err = readingsByDeviceNameAndResourceNamesAndTimeRange(conn, deviceName, resourceNames, start, end, offset, limit)
	if err != nil {
		return readings, errors.NewCommonEdgeX(errors.Kind(err),
			fmt.Sprintf("fail to query readings by deviceName %s, resourceNames %v and time range %v ~ %v", deviceName, resourceNames, start, end), err)
	}

	return readings, nil
}

func (c *Client) ReadingsByDeviceNameAndTimeRange(deviceName string, start int64, end int64, offset int, limit int) (readings []model.Reading, err errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	readings, err = readingsByDeviceNameAndTimeRange(conn, deviceName, start, end, offset, limit)
	if err != nil {
		return readings, errors.NewCommonEdgeX(errors.Kind(err),
			fmt.Sprintf("fail to query readings by deviceName %s, and time range %v ~ %v", deviceName, start, end), err)
	}

	return readings, nil
}

func (c *Client) ReadingCountByDeviceNameAndTimeRange(deviceName string, start int64, end int64) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberCountByScoreRange(conn, CreateKey(ReadingsCollectionDeviceName, deviceName), start, end)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// AddProvisionWatcher adds a new provision watcher
func (c *Client) AddProvisionWatcher(pw model.ProvisionWatcher) (model.ProvisionWatcher, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	if len(pw.Id) == 0 {
		pw.Id = uuid.New().String()
	}

	return addProvisionWatcher(conn, pw)
}

// ProvisionWatcherById gets a provision watcher by id
func (c *Client) ProvisionWatcherById(id string) (provisionWatcher model.ProvisionWatcher, edgexErr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	provisionWatcher, edgexErr = provisionWatcherById(conn, id)
	if edgexErr != nil {
		return provisionWatcher, errors.NewCommonEdgeX(errors.Kind(edgexErr), fmt.Sprintf("failed to query provision watcher by id %s", id), edgexErr)
	}

	return
}

// ProvisionWatcherByName gets a provision watcher by name
func (c *Client) ProvisionWatcherByName(name string) (provisionWatcher model.ProvisionWatcher, edgexErr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	provisionWatcher, edgexErr = provisionWatcherByName(conn, name)
	if edgexErr != nil {
		return provisionWatcher, errors.NewCommonEdgeXWrapper(edgexErr)
	}

	return
}

// ProvisionWatchersByServiceName query provision watchers by offset, limit and service name
func (c *Client) ProvisionWatchersByServiceName(offset int, limit int, name string) (provisionWatchers []model.ProvisionWatcher, edgexErr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	provisionWatchers, edgexErr = provisionWatchersByServiceName(conn, offset, limit, name)
	if edgexErr != nil {
		return provisionWatchers, errors.NewCommonEdgeX(errors.Kind(edgexErr),
			fmt.Sprintf("failed to query provision watcher by offset %d, limit %d and service name %s", offset, limit, name), edgexErr)
	}

	return
}

// ProvisionWatchersByProfileName query provision watchers by offset, limit and profile name
func (c *Client) ProvisionWatchersByProfileName(offset int, limit int, name string) (provisionWatchers []model.ProvisionWatcher, edgexErr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	provisionWatchers, edgexErr = provisionWatchersByProfileName(conn, offset, limit, name)
	if edgexErr != nil {
		return provisionWatchers, errors.NewCommonEdgeX(errors.Kind(edgexErr),
			fmt.Sprintf("failed to query provision watcher by offset %d, limit %d and profile name %s", offset, limit, name), edgexErr)
	}

	return
}

// AllProvisionWatchers query provision watchers with offset, limit and labels
func (c *Client) AllProvisionWatchers(offset int, limit int, labels []string) (provisionWatchers []model.ProvisionWatcher, edgexErr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	provisionWatchers, edgexErr = provisionWatchersByLabels(conn, offset, limit, labels)
	if edgexErr != nil {
		return provisionWatchers, errors.NewCommonEdgeXWrapper(edgexErr)
	}

	return
}

// DeleteProvisionWatcherByName deletes a provision watcher by name
func (c *Client) DeleteProvisionWatcherByName(name string) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	edgeXerr := deleteProvisionWatcherByName(conn, name)
	if edgeXerr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("failed to delete the provision watcher with name %s", name), edgeXerr)
	}

	return nil
}

// Update a provision watcher
func (c *Client) UpdateProvisionWatcher(pw model.ProvisionWatcher) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	return updateProvisionWatcher(conn, pw)
}

// DeviceProfileCountByLabels returns the total count of Device Profiles with labels specified.  If no label is specified, the total count of all device profiles will be returned.
func (c *Client) DeviceProfileCountByLabels(labels []string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberCountByLabels(conn, ZREVRANGE, DeviceProfileCollection, labels)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// DeviceProfileCountByManufacturer returns the count of Device Profiles associated with specified manufacturer
func (c *Client) DeviceProfileCountByManufacturer(manufacturer string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(DeviceProfileCollectionManufacturer, manufacturer))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// DeviceProfileCountByModel returns the count of Device Profiles associated with specified model
func (c *Client) DeviceProfileCountByModel(model string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(DeviceProfileCollectionModel, model))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// DeviceProfileCountByManufacturerAndModel returns the count of Device Profiles associated with specified manufacturer and model
func (c *Client) DeviceProfileCountByManufacturerAndModel(manufacturer, model string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	profiles, edgeXerr := deviceProfilesByManufacturerAndModel(conn, 0, -1, manufacturer, model)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return uint32(len(profiles)), nil
}

func (c *Client) InUseResourceCount() (uint32, errors.EdgeX) {
	c.loggingClient.Warn("InUseResourceCount function didn't implement")
	return 0, nil
}

// DeviceServiceCountByLabels returns the total count of Device Services with labels specified.  If no label is specified, the total count of all device services will be returned.
func (c *Client) DeviceServiceCountByLabels(labels []string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberCountByLabels(conn, ZREVRANGE, DeviceServiceCollection, labels)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// DeviceCountByLabels returns the total count of Devices with labels specified.  If no label is specified, the total count of all devices will be returned.
func (c *Client) DeviceCountByLabels(labels []string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberCountByLabels(conn, ZREVRANGE, DeviceCollection, labels)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// DeviceCountByProfileName returns the count of Devices associated with specified profile
func (c *Client) DeviceCountByProfileName(profileName string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(DeviceCollectionProfileName, profileName))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// DeviceCountByServiceName returns the count of Devices associated with specified service
func (c *Client) DeviceCountByServiceName(serviceName string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(DeviceCollectionServiceName, serviceName))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// ProvisionWatcherCountByLabels returns the total count of Provision Watchers with labels specified.  If no label is specified, the total count of all provision watchers will be returned.
func (c *Client) ProvisionWatcherCountByLabels(labels []string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberCountByLabels(conn, ZREVRANGE, ProvisionWatcherCollection, labels)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// ProvisionWatcherCountByServiceName returns the count of Provision Watcher associated with specified service
func (c *Client) ProvisionWatcherCountByServiceName(name string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(ProvisionWatcherCollectionServiceName, name))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// ProvisionWatcherCountByProfileName returns the count of Provision Watcher associated with specified profile
func (c *Client) ProvisionWatcherCountByProfileName(name string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(ProvisionWatcherCollectionProfileName, name))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// AddSubscription adds a new subscription
func (c *Client) AddSubscription(subscription model.Subscription) (model.Subscription, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	if len(subscription.Id) == 0 {
		subscription.Id = uuid.New().String()
	}

	return addSubscription(conn, subscription)
}

// AllSubscriptions returns multiple subscriptions per query criteria, including
// offset: The number of items to skip before starting to collect the result set.
// limit: The maximum number of items to return.
func (c *Client) AllSubscriptions(offset int, limit int) ([]model.Subscription, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	subscriptions, edgeXerr := allSubscriptions(conn, offset, limit)
	if edgeXerr != nil {
		return subscriptions, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return subscriptions, nil
}

// SubscriptionsByCategory queries subscriptions by offset, limit and category
func (c *Client) SubscriptionsByCategory(offset int, limit int, category string) (subscriptions []model.Subscription, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	subscriptions, edgeXerr = subscriptionsByCategory(conn, offset, limit, category)
	if edgeXerr != nil {
		return subscriptions, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query subscriptions by offset %d, limit %d and category %s", offset, limit, category), edgeXerr)
	}
	return subscriptions, nil
}

// SubscriptionsByLabel queries subscriptions by offset, limit and label
func (c *Client) SubscriptionsByLabel(offset int, limit int, label string) (subscriptions []model.Subscription, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	subscriptions, edgeXerr = subscriptionsByLabel(conn, offset, limit, label)
	if edgeXerr != nil {
		return subscriptions, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query subscriptions by offset %d, limit %d and label %s", offset, limit, label), edgeXerr)
	}
	return subscriptions, nil
}

// SubscriptionsByReceiver queries subscriptions by offset, limit and receiver
func (c *Client) SubscriptionsByReceiver(offset int, limit int, receiver string) (subscriptions []model.Subscription, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	subscriptions, edgeXerr = subscriptionsByReceiver(conn, offset, limit, receiver)
	if edgeXerr != nil {
		return subscriptions, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query subscriptions by offset %d, limit %d and receiver %s", offset, limit, receiver), edgeXerr)
	}
	return subscriptions, nil
}

// SubscriptionById gets a subscription by id
func (c *Client) SubscriptionById(id string) (subscription model.Subscription, edgexErr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	subscription, edgexErr = subscriptionById(conn, id)
	if edgexErr != nil {
		return subscription, errors.NewCommonEdgeX(errors.Kind(edgexErr), fmt.Sprintf("failed to query subscription by id %s", id), edgexErr)
	}

	return
}

// SubscriptionByName queries subscription by name
func (c *Client) SubscriptionByName(name string) (subscription model.Subscription, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	subscription, edgeXerr = subscriptionByName(conn, name)
	if edgeXerr != nil {
		return subscription, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query subscription by name %s", name), edgeXerr)
	}
	return subscription, nil
}

// UpdateSubscription updates a new subscription
func (c *Client) UpdateSubscription(subscription model.Subscription) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)
	return updateSubscription(conn, subscription)
}

// DeleteSubscriptionByName deletes a subscription by name
func (c *Client) DeleteSubscriptionByName(name string) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	edgeXerr := deleteSubscriptionByName(conn, name)
	if edgeXerr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to delete the subscription with name %s", name), edgeXerr)
	}

	return nil
}

// SubscriptionsByCategoriesAndLabels queries subscriptions by offset, limit, categories and labels
func (c *Client) SubscriptionsByCategoriesAndLabels(offset int, limit int, categories []string, labels []string) (subscriptions []model.Subscription, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	subscriptions, edgeXerr = subscriptionsByCategoriesAndLabels(conn, offset, limit, categories, labels)
	if edgeXerr != nil {
		return subscriptions, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query subscriptions by offset %d, limit %d, categories %v and labels %v", offset, limit, categories, labels), edgeXerr)
	}
	return subscriptions, nil
}

// AddNotification adds a new notification
func (c *Client) AddNotification(notification model.Notification) (model.Notification, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	if len(notification.Id) == 0 {
		notification.Id = uuid.New().String()
	}

	return addNotification(conn, notification)
}

// NotificationsByCategory queries notifications by offset, limit and category
func (c *Client) NotificationsByCategory(offset int, limit int, ack, category string) (notifications []model.Notification, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notifications, edgeXerr = notificationsByCategory(conn, offset, limit, ack, category)
	if edgeXerr != nil {
		return notifications, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query notifications by offset %d, limit %d, ack %s, and category %s", offset, limit, ack, category), edgeXerr)
	}
	return notifications, nil
}

// NotificationsByLabel queries notifications by offset, limit and label
func (c *Client) NotificationsByLabel(offset int, limit int, ack, label string) (notifications []model.Notification, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notifications, edgeXerr = notificationsByLabel(conn, offset, limit, ack, label)
	if edgeXerr != nil {
		return notifications, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query notifications by offset %d, limit %d, ack %s, and label %s", offset, limit, ack, label), edgeXerr)
	}
	return notifications, nil
}

// NotificationById gets a notification by id
func (c *Client) NotificationById(id string) (notification model.Notification, edgexErr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notification, edgexErr = notificationById(conn, id)
	if edgexErr != nil {
		return notification, errors.NewCommonEdgeX(errors.Kind(edgexErr), fmt.Sprintf("failed to query notification by id %s", id), edgexErr)
	}
	return
}

// NotificationsByStatus queries notifications by offset, limit and status
func (c *Client) NotificationsByStatus(offset int, limit int, ack, status string) (notifications []model.Notification, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notifications, edgeXerr = notificationsByStatus(conn, offset, limit, ack, status)
	if edgeXerr != nil {
		return notifications, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query notifications by offset %d, limit %d, ack %s, and status %s", offset, limit, ack, status), edgeXerr)
	}
	return notifications, nil
}

// NotificationsByTimeRange query notifications by time range, ack, offset, and limit
func (c *Client) NotificationsByTimeRange(start int64, end int64, offset int, limit int, ack string) (notifications []model.Notification, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notifications, edgeXerr = notificationsByTimeRange(conn, start, end, offset, limit, ack)
	if edgeXerr != nil {
		return notifications, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query notifications by time range %v ~ %v, offset %d, ack %s, and limit %d", start, end, offset, ack, limit), edgeXerr)
	}
	return notifications, nil
}

// NotificationsByCategoriesAndLabels queries notifications by ack, offset, limit, categories and labels
func (c *Client) NotificationsByCategoriesAndLabels(offset int, limit int, categories []string, labels []string, ack string) (notifications []model.Notification, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notifications, edgeXerr = notificationsByCategoriesAndLabels(conn, offset, limit, categories, labels, ack)
	if edgeXerr != nil {
		return notifications, errors.NewCommonEdgeX(errors.Kind(edgeXerr),
			fmt.Sprintf("fail to query notifications by offset %d, limit %d, categories %v, ack %s and labels %v", offset, limit, categories, ack, labels), edgeXerr)
	}
	return notifications, nil
}

// NotificationCountByCategory returns the count of Notification associated with specified category from the database
func (c *Client) NotificationCountByCategory(category, ack string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notifications, edgeXerr := notificationsByCategory(conn, 0, -1, ack, category)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return uint32(len(notifications)), nil
}

// NotificationCountByLabel returns the count of Notification associated with specified label from the database
func (c *Client) NotificationCountByLabel(label, ack string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notifications, edgeXerr := notificationsByLabel(conn, 0, -1, ack, label)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return uint32(len(notifications)), nil
}

// NotificationCountByStatus returns the count of Notification associated with specified status from the database
func (c *Client) NotificationCountByStatus(status, ack string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notifications, edgeXerr := notificationsByStatus(conn, 0, -1, ack, status)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return uint32(len(notifications)), nil
}

// NotificationCountByTimeRange returns the count of Notification from the database within specified time range
func (c *Client) NotificationCountByTimeRange(start int64, end int64, ack string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notifications, edgeXerr := notificationsByTimeRange(conn, start, end, 0, -1, ack)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return uint32(len(notifications)), nil
}

// NotificationCountByCategoriesAndLabels returns the count of Notification associated with specified categories and labels from the database
func (c *Client) NotificationCountByCategoriesAndLabels(categories []string, labels []string, ack string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notifications, edgeXerr := notificationsByCategoriesAndLabels(conn, 0, -1, categories, labels, ack)
	if edgeXerr != nil {
		return uint32(0), errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to query notifications by categories %v, labels %v, and ack %s", categories, labels, ack), edgeXerr)
	}
	return uint32(len(notifications)), nil
}

// NotificationCountByQueryConditions returns the count of Notification associated with specified condition from the database
func (c *Client) NotificationCountByQueryConditions(condition requests.NotificationQueryCondition, ack string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notifications, edgeXerr := notificationByQueryConditions(conn, 0, -1, condition, ack)
	if edgeXerr != nil {
		return uint32(0), errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to query notifications by condition %v and ack %s", condition, ack), edgeXerr)
	}
	return uint32(len(notifications)), nil
}

// NotificationsByQueryConditions queries notifications by offset, limit, categories and time range
func (c *Client) NotificationsByQueryConditions(offset int, limit int, condition requests.NotificationQueryCondition,
	ack string) (notifications []model.Notification, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notifications, edgeXerr = notificationByQueryConditions(conn, offset, limit, condition, ack)
	if edgeXerr != nil {
		return notifications, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return notifications, nil
}

// NotificationTotalCount returns the total count of Notification from the database
func (c *Client) NotificationTotalCount() (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, NotificationCollection)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// LatestNotificationByOffset returns a latest notification by offset
func (c *Client) LatestNotificationByOffset(offset uint32) (model.Notification, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notification, edgeXerr := latestNotificationByOffset(conn, int(offset))
	if edgeXerr != nil {
		return model.Notification{}, errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return notification, nil
}

// SubscriptionTotalCount returns the total count of Subscription from the database
func (c *Client) SubscriptionTotalCount() (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, SubscriptionCollection)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// SubscriptionCountByCategory returns the count of Subscription associated with specified category from the database
func (c *Client) SubscriptionCountByCategory(category string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(SubscriptionCollectionCategory, category))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// SubscriptionCountByLabel returns the count of Subscription associated with specified label from the database
func (c *Client) SubscriptionCountByLabel(label string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(SubscriptionCollectionLabel, label))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// SubscriptionCountByReceiver returns the count of Subscription associated with specified receiver from the database
func (c *Client) SubscriptionCountByReceiver(receiver string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(SubscriptionCollectionReceiver, receiver))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// TransmissionTotalCount returns the total count of Transmission from the database
func (c *Client) TransmissionTotalCount() (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, TransmissionCollection)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// TransmissionCountBySubscriptionName returns the count of Transmission associated with specified subscription name from the database
func (c *Client) TransmissionCountBySubscriptionName(subscriptionName string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(TransmissionCollectionSubscriptionName, subscriptionName))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// TransmissionCountByStatus returns the count of Transmission associated with specified status name from the database
func (c *Client) TransmissionCountByStatus(status string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(TransmissionCollectionStatus, status))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// TransmissionCountByTimeRange returns the count of Transmission from the database within specified time range
func (c *Client) TransmissionCountByTimeRange(start int64, end int64) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberCountByScoreRange(conn, TransmissionCollectionCreated, start, end)
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// DeleteNotificationById deletes a notification by id
func (c *Client) DeleteNotificationById(id string) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	edgeXerr := deleteNotificationById(conn, id)
	if edgeXerr != nil {
		return errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to delete the notification with id %s", id), edgeXerr)
	}

	return nil
}

// DeleteNotificationById deletes notifications by ids
func (c *Client) DeleteNotificationByIds(ids []string) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)
	edgeXerr := deleteNotificationByIds(conn, ids)
	if edgeXerr != nil {
		return errors.NewCommonEdgeXWrapper(edgeXerr)
	}
	return nil
}

// UpdateNotification updates a notification
func (c *Client) UpdateNotification(n model.Notification) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)
	return updateNotification(conn, n)
}

// UpdateNotificationAckStatusByIds bulk updates acknowledgement status
func (c *Client) UpdateNotificationAckStatusByIds(ack bool, ids []string) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	notifications, edgexErr := notificationByIds(conn, ids)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}
	return updateNotificationAckStatus(conn, ack, notifications)
}

// AddTransmission adds a new transmission
func (c *Client) AddTransmission(t model.Transmission) (model.Transmission, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	if len(t.Id) == 0 {
		t.Id = uuid.New().String()
	}

	return addTransmission(conn, t)
}

// UpdateTransmission updates a transmission
func (c *Client) UpdateTransmission(trans model.Transmission) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)
	return updateTransmission(conn, trans)
}

// TransmissionById gets a transmission by id
func (c *Client) TransmissionById(id string) (trans model.Transmission, edgexErr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	trans, edgexErr = transmissionById(conn, id)
	if edgexErr != nil {
		return trans, errors.NewCommonEdgeX(errors.Kind(edgexErr), fmt.Sprintf("failed to query transmission by id %s", id), edgexErr)
	}
	return
}

// TransmissionsByTimeRange query transmissions by time range, offset, and limit
func (c *Client) TransmissionsByTimeRange(start int64, end int64, offset int, limit int) (transmissions []model.Transmission, err errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	transmissions, err = transmissionsByTimeRange(conn, start, end, offset, limit)
	if err != nil {
		return transmissions, errors.NewCommonEdgeX(errors.Kind(err),
			fmt.Sprintf("fail to query transmissions by time range %v ~ %v, offset %d, and limit %d", start, end, offset, limit), err)
	}
	return transmissions, nil
}

// AllTransmissions returns multiple transmissions per query criteria, including
// offset: The number of items to skip before starting to collect the result set.
// limit: The maximum number of items to return.
func (c *Client) AllTransmissions(offset int, limit int) ([]model.Transmission, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	transmission, err := allTransmissions(conn, offset, limit)
	if err != nil {
		return transmission, errors.NewCommonEdgeXWrapper(err)
	}
	return transmission, nil
}

// TransmissionsByStatus queries transmissions by offset, limit and status
func (c *Client) TransmissionsByStatus(offset int, limit int, status string) (transmissions []model.Transmission, err errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	transmissions, err = transmissionsByStatus(conn, offset, limit, status)
	if err != nil {
		return transmissions, errors.NewCommonEdgeX(errors.Kind(err),
			fmt.Sprintf("fail to query transmissions by offset %d, limit %d and status %s", offset, limit, status), err)
	}
	return transmissions, nil
}

// TransmissionsBySubscriptionName queries transmissions by offset, limit and subscription name
func (c *Client) TransmissionsBySubscriptionName(offset int, limit int, subscriptionName string) (transmissions []model.Transmission, err errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	transmissions, err = transmissionsBySubscriptionName(conn, offset, limit, subscriptionName)
	if err != nil {
		return transmissions, errors.NewCommonEdgeX(errors.Kind(err),
			fmt.Sprintf("fail to query transmissions by offset %d, limit %d and subscription name %s", offset, limit, subscriptionName), err)
	}
	return transmissions, nil
}

// TransmissionsByNotificationId queries transmissions by offset, limit and notification id
func (c *Client) TransmissionsByNotificationId(offset int, limit int, id string) (transmissions []model.Transmission, err errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	transmissions, err = transmissionsByNotificationId(conn, offset, limit, id)
	if err != nil {
		return transmissions, errors.NewCommonEdgeX(errors.Kind(err),
			fmt.Sprintf("fail to query transmissions by offset %d, limit %d and notification id %s", offset, limit, id), err)
	}
	return transmissions, nil
}

// TransmissionCountByNotificationId returns the count of Transmission associated with specified notification id from the database
func (c *Client) TransmissionCountByNotificationId(id string) (uint32, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	count, edgeXerr := getMemberNumber(conn, ZCARD, CreateKey(TransmissionCollectionNotificationId, id))
	if edgeXerr != nil {
		return 0, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return count, nil
}

// LatestReadingByOffset returns a latest reading by offset
func (c *Client) LatestReadingByOffset(offset uint32) (model.Reading, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	reading, edgeXerr := latestReadingByOffset(conn, int(offset))
	if edgeXerr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	return reading, nil
}

// KeeperKeys returns the values stored for the specified key or with the same key prefix
func (c *Client) KeeperKeys(key string, keyOnly bool, isRaw bool) (kvs []model.KVResponse, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	key = replaceKeyDelimiterForDB(key)
	kvs, edgeXerr = keeperKeys(conn, key, keyOnly, isRaw)
	if edgeXerr != nil {
		// replace the key delimiter from colon(:) to slash(/)
		errRespKey := replaceKeyDelimiterForKeeper(key)
		return kvs, errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to get key %s", errRespKey), edgeXerr)
	}

	for _, kv := range kvs {
		// check if the KVResponse interface is KV struct
		if convertedKV, ok := kv.(*model.KVS); ok {
			// convert KVResponse interface to KV struct and replace the key delimiter
			convertedKV.SetKey(replaceKeyDelimiterForKeeper(convertedKV.Key))
		} else {
			// dereference the KeyOnly pointer and convert to string
			originKey := string(*kv.(*model.KeyOnly))
			// convert KVResponse interface to KeyOnly struct and replace the key delimiter
			kv.(*model.KeyOnly).SetKey(replaceKeyDelimiterForKeeper(originKey))
		}
	}
	return kvs, nil
}

// AddKeeperKeys returns the values stored for the specified key or with the same key prefix
func (c *Client) AddKeeperKeys(kv model.KVS, isFlatten bool) (keys []model.KeyOnly, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	kv.Key = replaceKeyDelimiterForDB(kv.Key)
	keys, edgeXerr = addKeeperKeys(conn, kv, isFlatten)
	if edgeXerr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgeXerr)
	}

	// replace the key delimiter in the response for Keeper
	for i, key := range keys {
		// convert the KeyOnly type to string
		originKey := string(key)

		// replace the key delimiter from colon(:) to slash(/)
		keys[i].SetKey(replaceKeyDelimiterForKeeper(originKey))
	}
	return keys, nil
}

// DeleteKeeperKeys delete the specified key or keys with the same prefix
func (c *Client) DeleteKeeperKeys(key string, prefixMatch bool) (kvs []model.KeyOnly, edgeXerr errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	key = replaceKeyDelimiterForDB(key)
	kvs, edgeXerr = deleteKeeperKeys(conn, key, prefixMatch)
	if edgeXerr != nil {
		return kvs, errors.NewCommonEdgeX(errors.Kind(edgeXerr), fmt.Sprintf("fail to delete key %s", replaceKeyDelimiterForKeeper(key)), edgeXerr)
	}

	for i, kv := range kvs {
		// convert the KeyOnly type to string
		originKey := string(kv)

		// replace the key delimiter from colon(:) to slash(/)
		kvs[i].SetKey(replaceKeyDelimiterForKeeper(originKey))
	}
	return kvs, nil
}

func (c *Client) AddRegistration(r model.Registration) (model.Registration, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	r, err := addRegistration(conn, r)
	if err != nil {
		return model.Registration{}, errors.NewCommonEdgeXWrapper(err)
	}

	return r, nil
}

func (c *Client) DeleteRegistrationByServiceId(id string) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	err := deleteRegistrationByServiceId(conn, id)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return nil
}

func (c *Client) Registrations() ([]model.Registration, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	registries, err := registrations(conn, 0, -1)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	return registries, nil
}

func (c *Client) RegistrationByServiceId(id string) (model.Registration, errors.EdgeX) {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	r, err := registrationById(conn, id)
	if err != nil {
		return model.Registration{}, errors.NewCommonEdgeXWrapper(err)
	}

	return r, nil
}

func (c *Client) UpdateRegistration(r model.Registration) errors.EdgeX {
	conn := c.Pool.Get()
	defer closeRedisConnection(conn, c.loggingClient)

	err := updateRegistration(conn, r)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return nil
}
