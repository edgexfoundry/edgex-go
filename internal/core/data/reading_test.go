package data

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/stretchr/testify/mock"
)

func newReadingsMockDB() interfaces.DBClient {
	db := &dbMock.DBClient{}

	db.On("Readings").Return(buildReadings(), nil)

	return db
}

func TestGetAllReadings(t *testing.T) {
	reset()
	Configuration.Service.MaxResultCount = 5

	dbClient = newReadingsMockDB()

	_, err := getAllReadings(logger.NewMockClient())

	if err != nil {
		t.Errorf("Unexpected error thrown getting all readings: %s", err.Error())
	}
}

func TestGetAllReadingsOverLimit(t *testing.T) {
	reset()
	Configuration.Service.MaxResultCount = 1

	dbClient = newReadingsMockDB()

	_, err := getAllReadings(logger.NewMockClient())

	if err != nil {
		switch err.(type) {
		case errors.ErrLimitExceeded:
			return
		default:
			t.Errorf("Unexpected error getting all readings")
		}
	}

	if err == nil {
		t.Errorf("Expected error getting all readings")
	}
}

func TestGetAllReadingsError(t *testing.T) {
	reset()
	Configuration.Service.MaxResultCount = 5
	myMock := &dbMock.DBClient{}

	myMock.On("Readings").Return([]models.Reading{}, fmt.Errorf("some error"))

	dbClient = myMock

	_, err := getAllReadings(logger.NewMockClient())

	if err == nil {
		t.Errorf("Expected error getting all readings")
	}
}

func TestAddReading(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("AddReading", mock.Anything).Return("", nil)

	dbClient = myMock

	_, err := addReading(models.Reading{Name: "valid"}, logger.NewMockClient())

	if err != nil {
		t.Errorf("Unexpected error adding reading")
	}
}

func TestAddReadingError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("AddReading", mock.Anything).Return("", fmt.Errorf("some error"))

	dbClient = myMock

	_, err := addReading(models.Reading{}, logger.NewMockClient())

	if err == nil {
		t.Errorf("Expected error adding reading")
	}
}

func TestGetReadingById(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingById", mock.Anything).Return(models.Reading{}, nil)

	dbClient = myMock

	_, err := getReadingById("valid", logger.NewMockClient())

	if err != nil {
		t.Errorf("Unexpected error getting reading by ID")
	}
}

func TestGetReadingByIdNotFound(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingById", mock.Anything).Return(models.Reading{}, db.ErrNotFound)

	dbClient = myMock

	_, err := getReadingById("404", logger.NewMockClient())

	if err != nil {
		switch err.(type) {
		case errors.ErrDbNotFound:
			return
		default:
			t.Errorf("Unexpected error getting reading by ID missing in DB")
		}
	}

	if err == nil {
		t.Errorf("Expected error getting reading by ID missing in DB")
	}
}

func TestGetReadingByIdError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingById", mock.Anything).Return(models.Reading{}, fmt.Errorf("some error"))

	dbClient = myMock

	_, err := getReadingById("error", logger.NewMockClient())

	if err == nil {
		t.Errorf("Expected error getting reading by ID with some error")
	}
}

func TestDeleteReadingById(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("DeleteReadingById", mock.Anything).Return(nil).Once()

	dbClient = myMock

	err := deleteReadingById("valid", logger.NewMockClient())

	if err != nil {
		t.Errorf("Unexpected error deleting reading by ID")
	}

	myMock.AssertExpectations(t)
}

func TestDeleteReadingByIdError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("DeleteReadingById", mock.Anything).Return(fmt.Errorf("some error"))

	dbClient = myMock

	err := deleteReadingById("invalid", logger.NewMockClient())

	if err == nil {
		t.Errorf("Expected error deleting reading by ID")
	}

	myMock.AssertExpectations(t)
}

func TestCountReadings(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingCount").Return(2, nil)

	dbClient = myMock

	_, err := countReadings(logger.NewMockClient())

	if err != nil {
		t.Errorf("Unexpected error in CountReadings")
	}
}

func TestCountReadingsError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingCount").Return(2, fmt.Errorf("some error"))

	dbClient = myMock

	_, err := countReadings(logger.NewMockClient())

	if err == nil {
		t.Errorf("Expected error in CountReadings")
	}
}

func TestGetReadingsByDevice(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByDevice", mock.Anything, mock.Anything).Return(buildReadings(), nil)

	dbClient = myMock

	expectedReadings, err := getReadingsByDevice("valid", 0, context.Background(), logger.NewMockClient())

	if err != nil {
		t.Errorf("Unexpected error in getReadingsByDevice")
	}

	if len(buildReadings()) != len(expectedReadings) {
		t.Errorf("Found %d readings, expected %d", len(expectedReadings), len(buildReadings()))
	}
}

func TestGetReadingsByDeviceError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByDevice", mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))

	dbClient = myMock

	_, err := getReadingsByDevice("error", 0, context.Background(), logger.NewMockClient())

	if err == nil {
		t.Errorf("Expected error in getReadingsByDevice")
	}
}

func TestGetReadingsByValueDescriptor(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByValueDescriptor", mock.Anything, mock.Anything).Return([]models.Reading{}, nil)

	dbClient = myMock

	_, err := getReadingsByValueDescriptor("valid", 0, logger.NewMockClient())

	if err != nil {
		t.Errorf("Unexpected error getting readings by value descriptor")
	}
}

func TestGetReadingsByValueDescriptorOverLimit(t *testing.T) {
	reset()
	dbClient = nil

	_, err := getReadingsByValueDescriptor("", math.MaxInt32, logger.NewMockClient())

	if err == nil {
		t.Errorf("Expected error getting readings by value descriptor")
	}
}

func TestGetReadingsByValueDescriptorError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByValueDescriptor", mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))

	dbClient = myMock

	_, err := getReadingsByValueDescriptor("error", 0, logger.NewMockClient())

	if err == nil {
		t.Errorf("Expected error in getting readings by value descriptor")
	}
}

func TestGetReadingsByValueDescriptorNames(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByValueDescriptorNames", mock.Anything, mock.Anything).Return([]models.Reading{}, nil)

	dbClient = myMock

	_, err := getReadingsByValueDescriptorNames([]string{"valid"}, 0, logger.NewMockClient())

	if err != nil {
		t.Errorf("Unexpected error getting readings by value descriptor names")
	}
}

func TestGetReadingsByValueDescriptorNamesError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByValueDescriptorNames", mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))

	dbClient = myMock

	_, err := getReadingsByValueDescriptorNames([]string{"error"}, 0, logger.NewMockClient())

	if err == nil {
		t.Errorf("Expected error in getting readings by value descriptor names")
	}
}

func TestGetReadingsByCreationTime(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByCreationTime", mock.Anything, mock.Anything, mock.Anything).Return([]models.Reading{}, nil)

	dbClient = myMock

	_, err := getReadingsByCreationTime(0xBEEF, 0, 0, logger.NewMockClient())

	if err != nil {
		t.Errorf("Unexpected error getting readings by creation time")
	}
}

func TestGetReadingsByCreationTimeError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByCreationTime", mock.Anything, mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))

	dbClient = myMock

	_, err := getReadingsByCreationTime(0xDEADBEEF, 0, 0, logger.NewMockClient())

	if err == nil {
		t.Errorf("Expected error in getting readings by creation time")
	}
}

func TestGetReadingsByDeviceAndValueDescriptor(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByDeviceAndValueDescriptor", mock.Anything, mock.Anything, mock.Anything).Return([]models.Reading{}, nil)

	dbClient = myMock

	_, err := getReadingsByDeviceAndValueDescriptor("valid", "valid", 0, logger.NewMockClient())

	if err != nil {
		t.Errorf("Unexpected error getting readings by device and value descriptor")
	}
}

func TestGetReadingsByDeviceAndValueDescriptorError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByDeviceAndValueDescriptor", mock.Anything, mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))

	dbClient = myMock

	_, err := getReadingsByDeviceAndValueDescriptor("error", "error", 0, logger.NewMockClient())

	if err == nil {
		t.Errorf("Expected error in getting readings by device and value descriptor")
	}
}
