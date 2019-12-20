package data

import (
	"context"
	"fmt"
	"math"
	"testing"

	dataConfig "github.com/edgexfoundry/edgex-go/internal/core/data/config"
	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/interfaces/mocks"
	dataMocks "github.com/edgexfoundry/edgex-go/internal/core/data/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/edgexfoundry/go-mod-bootstrap/config"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
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
	_, err := getAllReadings(
		logger.NewMockClient(),
		newReadingsMockDB(),
		&dataConfig.ConfigurationStruct{
			Service: config.ServiceInfo{
				MaxResultCount: 5,
			},
		})

	if err != nil {
		t.Errorf("Unexpected error thrown getting all readings: %s", err.Error())
	}
}

func TestGetAllReadingsOverLimit(t *testing.T) {
	reset()
	_, err := getAllReadings(logger.NewMockClient(), newReadingsMockDB(), &dataConfig.ConfigurationStruct{
		Service: config.ServiceInfo{
			MaxResultCount: 1,
		}})

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
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("Readings").Return([]models.Reading{}, fmt.Errorf("some error"))
	_, err := getAllReadings(logger.NewMockClient(), dbClientMock, &dataConfig.ConfigurationStruct{
		Service: config.ServiceInfo{
			MaxResultCount: 5,
		}})

	if err == nil {
		t.Errorf("Expected error getting all readings")
	}
}

func TestAddReading(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AddReading", mock.Anything).Return("", nil)
	_, err := addReading(models.Reading{Name: "valid"}, logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error adding reading")
	}
}

func TestAddReadingError(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("AddReading", mock.Anything).Return("", fmt.Errorf("some error"))
	_, err := addReading(models.Reading{}, logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error adding reading")
	}
}

func TestGetReadingById(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingById", mock.Anything).Return(models.Reading{}, nil)
	_, err := getReadingById("valid", logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error getting reading by ID")
	}
}

func TestGetReadingByIdNotFound(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingById", mock.Anything).Return(models.Reading{}, db.ErrNotFound)
	_, err := getReadingById("404", logger.NewMockClient(), dbClientMock)
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
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingById", mock.Anything).Return(models.Reading{}, fmt.Errorf("some error"))
	_, err := getReadingById("error", logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error getting reading by ID with some error")
	}
}

func TestDeleteReadingById(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeleteReadingById", mock.Anything).Return(nil).Once()
	err := deleteReadingById("valid", logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error deleting reading by ID")
	}

	dbClientMock.AssertExpectations(t)
}

func TestDeleteReadingByIdError(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("DeleteReadingById", mock.Anything).Return(fmt.Errorf("some error"))
	err := deleteReadingById("invalid", logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error deleting reading by ID")
	}

	dbClientMock.AssertExpectations(t)
}

func TestCountReadings(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingCount").Return(2, nil)
	_, err := countReadings(logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error in CountReadings")
	}
}

func TestCountReadingsError(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingCount").Return(2, fmt.Errorf("some error"))
	_, err := countReadings(logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error in CountReadings")
	}
}

func TestGetReadingsByDevice(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByDevice", mock.Anything, mock.Anything).Return(buildReadings(), nil)
	expectedReadings, err := getReadingsByDevice("valid", 0, context.Background(), logger.NewMockClient(), dbClientMock, dataMocks.NewMockDeviceClient(), &dataConfig.ConfigurationStruct{})
	if err != nil {
		t.Errorf("Unexpected error in getReadingsByDevice")
	}

	if len(buildReadings()) != len(expectedReadings) {
		t.Errorf("Found %d readings, expected %d", len(expectedReadings), len(buildReadings()))
	}
}

func TestGetReadingsByDeviceError(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByDevice", mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))
	_, err := getReadingsByDevice("error", 0, context.Background(), logger.NewMockClient(), dbClientMock, dataMocks.NewMockDeviceClient(), &dataConfig.ConfigurationStruct{})
	if err == nil {
		t.Errorf("Expected error in getReadingsByDevice")
	}
}

func TestGetReadingsByValueDescriptor(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByValueDescriptor", mock.Anything, mock.Anything).Return([]models.Reading{}, nil)
	_, err := getReadingsByValueDescriptor("valid", 0, logger.NewMockClient(), dbClientMock, &dataConfig.ConfigurationStruct{})
	if err != nil {
		t.Errorf("Unexpected error getting readings by value descriptor")
	}
}

func TestGetReadingsByValueDescriptorOverLimit(t *testing.T) {
	reset()
	_, err := getReadingsByValueDescriptor("", math.MaxInt32, logger.NewMockClient(), nil, &dataConfig.ConfigurationStruct{})
	if err == nil {
		t.Errorf("Expected error getting readings by value descriptor")
	}
}

func TestGetReadingsByValueDescriptorError(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByValueDescriptor", mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))
	_, err := getReadingsByValueDescriptor("error", 0, logger.NewMockClient(), dbClientMock, &dataConfig.ConfigurationStruct{})
	if err == nil {
		t.Errorf("Expected error in getting readings by value descriptor")
	}
}

func TestGetReadingsByValueDescriptorNames(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByValueDescriptorNames", mock.Anything, mock.Anything).Return([]models.Reading{}, nil)
	_, err := getReadingsByValueDescriptorNames([]string{"valid"}, 0, logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error getting readings by value descriptor names")
	}
}

func TestGetReadingsByValueDescriptorNamesError(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByValueDescriptorNames", mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))
	_, err := getReadingsByValueDescriptorNames([]string{"error"}, 0, logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error in getting readings by value descriptor names")
	}
}

func TestGetReadingsByCreationTime(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByCreationTime", mock.Anything, mock.Anything, mock.Anything).Return([]models.Reading{}, nil)
	_, err := getReadingsByCreationTime(0xBEEF, 0, 0, logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error getting readings by creation time")
	}
}

func TestGetReadingsByCreationTimeError(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByCreationTime", mock.Anything, mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))
	_, err := getReadingsByCreationTime(0xDEADBEEF, 0, 0, logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error in getting readings by creation time")
	}
}

func TestGetReadingsByDeviceAndValueDescriptor(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByDeviceAndValueDescriptor", mock.Anything, mock.Anything, mock.Anything).Return([]models.Reading{}, nil)
	_, err := getReadingsByDeviceAndValueDescriptor("valid", "valid", 0, logger.NewMockClient(), dbClientMock)
	if err != nil {
		t.Errorf("Unexpected error getting readings by device and value descriptor")
	}
}

func TestGetReadingsByDeviceAndValueDescriptorError(t *testing.T) {
	reset()
	dbClientMock := &dbMock.DBClient{}
	dbClientMock.On("ReadingsByDeviceAndValueDescriptor", mock.Anything, mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))
	_, err := getReadingsByDeviceAndValueDescriptor("error", "error", 0, logger.NewMockClient(), dbClientMock)
	if err == nil {
		t.Errorf("Expected error in getting readings by device and value descriptor")
	}
}
