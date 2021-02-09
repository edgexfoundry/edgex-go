package application

import (
	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// ReadingTotalCount return the count of all of readings currently stored in the database and error if any
func ReadingTotalCount(dic *di.Container) (uint32, errors.EdgeX) {
	dbClient := v2DataContainer.DBClientFrom(dic.Get)

	count, err := dbClient.ReadingTotalCount()
	if err != nil {
		return 0, errors.NewCommonEdgeXWrapper(err)
	}

	return count, nil
}

// AllReadings query events by offset, and limit
func AllReadings(offset int, limit int, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	dbClient := v2DataContainer.DBClientFrom(dic.Get)
	readingModels, err := dbClient.AllReadings(offset, limit)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}
	return convertReadingModelsToDTOs(readingModels)
}

// ReadingsByResourceName query readings with offset, limit, and resource name
func ReadingsByResourceName(offset int, limit int, resourceName string, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	if resourceName == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "resourceName is empty", nil)
	}
	dbClient := v2DataContainer.DBClientFrom(dic.Get)
	readingModels, err := dbClient.ReadingsByResourceName(offset, limit, resourceName)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}
	readings = make([]dtos.BaseReading, len(readingModels))
	for i, r := range readingModels {
		readings[i] = dtos.FromReadingModelToDTO(r)
	}
	return readings, nil
}

// ReadingsByDeviceName query readings with offset, limit, and device name
func ReadingsByDeviceName(offset int, limit int, name string, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	if name == "" {
		return readings, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2DataContainer.DBClientFrom(dic.Get)
	readingModels, err := dbClient.ReadingsByDeviceName(offset, limit, name)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}
	return convertReadingModelsToDTOs(readingModels)
}

// ReadingsByTimeRange query readings with offset, limit and time range
func ReadingsByTimeRange(start int, end int, offset int, limit int, dic *di.Container) (readings []dtos.BaseReading, err errors.EdgeX) {
	dbClient := v2DataContainer.DBClientFrom(dic.Get)
	readingModels, err := dbClient.ReadingsByTimeRange(start, end, offset, limit)
	if err != nil {
		return readings, errors.NewCommonEdgeXWrapper(err)
	}
	return convertReadingModelsToDTOs(readingModels)
}

func convertReadingModelsToDTOs(readingModels []models.Reading) (readings []dtos.BaseReading, err errors.EdgeX) {
	readings = make([]dtos.BaseReading, len(readingModels))
	for i, r := range readingModels {
		readings[i] = dtos.FromReadingModelToDTO(r)
	}
	return readings, nil
}

// ReadingCountByDeviceName return the count of all of readings associated with given device and error if any
func ReadingCountByDeviceName(deviceName string, dic *di.Container) (uint32, errors.EdgeX) {
	if deviceName == "" {
		return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, "name is empty", nil)
	}
	dbClient := v2DataContainer.DBClientFrom(dic.Get)
	count, err := dbClient.ReadingCountByDeviceName(deviceName)
	if err != nil {
		return 0, errors.NewCommonEdgeXWrapper(err)
	}

	return count, nil
}
