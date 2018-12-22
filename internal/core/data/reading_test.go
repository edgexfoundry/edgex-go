package data

import (
	"fmt"
	"math"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/data/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"

	"github.com/edgexfoundry/edgex-go/pkg/models"

	"github.com/stretchr/testify/mock"
)

func newReadingsMockDB() interfaces.DBClient {
	db := &dbMock.DBClient{}

	db.On("Readings").Return(buildReadings(), nil)

	return db
}

func TestGetAllReadings(t *testing.T) {
	reset()
	Configuration.Service.ReadMaxLimit = 5

	dbClient = newReadingsMockDB()

	_, err := getAllReadings()

	if err != nil {
		t.Errorf("Unexpected error thrown getting all readings: %s", err.Error())
	}
}

func TestGetAllReadingsOverLimit(t *testing.T) {
	reset()
	Configuration.Service.ReadMaxLimit = 1

	dbClient = newReadingsMockDB()

	_, err := getAllReadings()

	if err != nil {
		switch err.(type) {
		case *errors.ErrLimitExceeded:
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
	Configuration.Service.ReadMaxLimit = 5
	myMock := &dbMock.DBClient{}

	myMock.On("Readings").Return([]models.Reading{}, fmt.Errorf("some error"))

	dbClient = myMock

	_, err := getAllReadings()

	if err == nil {
		t.Errorf("Expected error getting all readings")
	}
}

func TestAddReading(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("AddReading", mock.Anything).Return("", nil)

	dbClient = myMock

	_, err := addReading(models.Reading{Name: "valid"})

	if err != nil {
		t.Errorf("Unexpected error adding reading")
	}
}

func TestAddReadingError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("AddReading", mock.Anything).Return("", fmt.Errorf("some error"))

	dbClient = myMock

	_, err := addReading(models.Reading{})

	if err == nil {
		t.Errorf("Expected error adding reading")
	}
}

func TestGetReadingById(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingById", mock.Anything).Return(models.Reading{}, nil)

	dbClient = myMock

	_, err := getReadingById("valid")

	if err != nil {
		t.Errorf("Unexpected error getting reading by ID")
	}
}

func TestGetReadingByIdNotFound(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingById", mock.Anything).Return(models.Reading{}, db.ErrNotFound)

	dbClient = myMock

	_, err := getReadingById("404")

	if err != nil {
		switch err.(type) {
		case *errors.ErrDbNotFound:
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

	_, err := getReadingById("error")

	if err == nil {
		t.Errorf("Expected error getting reading by ID with some error")
	}
}

func TestDeleteReadingById(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("DeleteReadingById", mock.Anything).Return(nil).Once()

	dbClient = myMock

	err := deleteReadingById("valid")

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

	err := deleteReadingById("invalid")

	if err == nil {
		t.Errorf("Expected error deleting reading by ID")
	}

	myMock.AssertExpectations(t)
}

func TestGetReadingsByDeviceId(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("EventsForDevice", mock.Anything).Return([]models.Event{{Readings: append(buildReadings(), buildReadings()...)}}, nil)

	dbClient = myMock

	expectedReadings, expectedNil := getReadingsByDeviceId(math.MaxInt32, "valid", "Pressure")

	if expectedReadings == nil {
		t.Errorf("Should return Readings")
	}

	if expectedNil != nil {
		t.Errorf("Should not throw error")
	}

	if len(expectedReadings) != len(buildReadings()) {
		t.Errorf("Returned %d readings, expected %d", len(expectedReadings), len(buildReadings()))
	}
}

func TestGetReadingsByDeviceIdLimited(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("EventsForDevice", mock.Anything).Return([]models.Event{{Readings: append(buildReadings(), buildReadings()...)}}, nil)

	dbClient = myMock

	for limit := 0; limit < 5; limit++ {
		expectedReadings, expectedNil := getReadingsByDeviceId(limit, "valid", "Pressure")

		if limit == 0 {
			if expectedReadings != nil {
				t.Errorf("Should return nil slice for zero limit")
			}
		} else if expectedReadings == nil {
			t.Errorf("Should return Readings, limit: %d", limit)
		}

		if len(expectedReadings) > limit {
			t.Errorf("Should only return %d Readings", limit)
		}

		if expectedNil != nil {
			t.Errorf("Should not throw error")
		}
	}
}

func TestGetReadingsByDeviceIdDBThrowsError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("EventsForDevice", mock.Anything).Return(nil, fmt.Errorf("some error"))

	dbClient = myMock

	expectedNil, expectedErr := getReadingsByDeviceId(0, "error", "")

	if expectedNil != nil {
		t.Errorf("Should not return Readings on error")
	}

	if expectedErr == nil {
		t.Errorf("Should throw error")
	}
}

func TestCountReadings(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingCount").Return(2, nil)

	dbClient = myMock

	_, err := countReadings()

	if err != nil {
		t.Errorf("Unexpected error in CountReadings")
	}
}

func TestCountReadingsError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingCount").Return(2, fmt.Errorf("some error"))

	dbClient = myMock

	_, err := countReadings()

	if err == nil {
		t.Errorf("Expected error in CountReadings")
	}
}

func TestGetReadingsByDevice(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByDevice", mock.Anything, mock.Anything).Return(buildReadings(), nil)

	dbClient = myMock

	expectedReadings, err := getReadingsByDevice("valid", 0)

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

	_, err := getReadingsByDevice("error", 0)

	if err == nil {
		t.Errorf("Expected error in getReadingsByDevice")
	}
}

func TestGetReadingsByValueDescriptor(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByValueDescriptor", mock.Anything, mock.Anything).Return([]models.Reading{}, nil)

	dbClient = myMock

	_, err := getReadingsByValueDescriptor("valid", 0)

	if err != nil {
		t.Errorf("Unexpected error getting readings by value descriptor")
	}
}

func TestGetReadingsByValueDescriptorOverLimit(t *testing.T) {
	reset()
	dbClient = nil

	_, err := getReadingsByValueDescriptor("", math.MaxInt32)

	if err == nil {
		t.Errorf("Expected error getting readings by value descriptor")
	}
}

func TestGetReadingsByValueDescriptorError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByValueDescriptor", mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))

	dbClient = myMock

	_, err := getReadingsByValueDescriptor("error", 0)

	if err == nil {
		t.Errorf("Expected error in getting readings by value descriptor")
	}
}

func TestGetReadingsByValueDescriptorNames(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByValueDescriptorNames", mock.Anything, mock.Anything).Return([]models.Reading{}, nil)

	dbClient = myMock

	_, err := getReadingsByValueDescriptorNames([]string{"valid"}, 0)

	if err != nil {
		t.Errorf("Unexpected error getting readings by value descriptor names")
	}
}

func TestGetReadingsByValueDescriptorNamesError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByValueDescriptorNames", mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))

	dbClient = myMock

	_, err := getReadingsByValueDescriptorNames([]string{"error"}, 0)

	if err == nil {
		t.Errorf("Expected error in getting readings by value descriptor names")
	}
}

func TestGetReadingsByCreationTime(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByCreationTime", mock.Anything, mock.Anything, mock.Anything).Return([]models.Reading{}, nil)

	dbClient = myMock

	_, err := getReadingsByCreationTime(0xBEEF, 0, 0)

	if err != nil {
		t.Errorf("Unexpected error getting readings by creation time")
	}
}

func TestGetReadingsByCreationTimeError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByCreationTime", mock.Anything, mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))

	dbClient = myMock

	_, err := getReadingsByCreationTime(0xDEADBEEF, 0, 0)

	if err == nil {
		t.Errorf("Expected error in getting readings by creation time")
	}
}

func TestGetReadingsByDeviceAndValueDescriptor(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByDeviceAndValueDescriptor", mock.Anything, mock.Anything, mock.Anything).Return([]models.Reading{}, nil)

	dbClient = myMock

	_, err := getReadingsByDeviceAndValueDescriptor("valid", "valid", 0)

	if err != nil {
		t.Errorf("Unexpected error getting readings by device and value descriptor")
	}
}

func TestGetReadingsByDeviceAndValueDescriptorError(t *testing.T) {
	reset()
	myMock := &dbMock.DBClient{}

	myMock.On("ReadingsByDeviceAndValueDescriptor", mock.Anything, mock.Anything, mock.Anything).Return([]models.Reading{}, fmt.Errorf("some error"))

	dbClient = myMock

	_, err := getReadingsByDeviceAndValueDescriptor("error", "error", 0)

	if err == nil {
		t.Errorf("Expected error in getting readings by device and value descriptor")
	}
}
