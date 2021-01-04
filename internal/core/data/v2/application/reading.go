package application

import (
	v2DataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
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
