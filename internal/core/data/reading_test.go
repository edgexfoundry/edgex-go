package data

import (
	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"math"
	"testing"
)

func TestGetAllReadings(t *testing.T) {
	reset()
	Configuration.Service.ReadMaxLimit = 5
	// memdb used in this test on purpose

	_, err := getAllReadings()

	if err != nil {
		t.Errorf("Unexpected error thrown getting all readings")
	}
}

func TestGetAllReadingsOverLimit(t *testing.T) {
	reset()
	Configuration.Service.ReadMaxLimit = 1
	// memdb used in this test on purpose

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
	dbClient = newMockDb()

	_, err := getAllReadings()

	if err == nil {
		t.Errorf("Expected error getting all readings")
	}
}

func TestAddReading(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := addReading(models.Reading{Name: "valid"})

	if err != nil {
		t.Errorf("Unexpected error adding reading")
	}
}

func TestAddReadingError(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := addReading(models.Reading{})

	if err == nil {
		t.Errorf("Expected error adding reading")
	}
}

func TestGetReadingById(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getReadingById("valid")

	if err != nil {
		t.Errorf("Unexpected error getting reading by ID")
	}
}

func TestGetReadingByIdNotFound(t *testing.T) {
	reset()
	dbClient = newMockDb()

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
	dbClient = newMockDb()

	_, err := getReadingById("error")

	if err == nil {
		t.Errorf("Expected error getting reading by ID with some error")
	}
}

func TestDeleteReadingById(t *testing.T) {
	reset()
	dbClient = newMockDb()

	err := deleteReadingById("valid")

	if err != nil {
		t.Errorf("Unexpected error deleting reading by ID")
	}
}

func TestDeleteReadingByIdError(t *testing.T) {
	reset()
	dbClient = newMockDb()

	err := deleteReadingById("invalid")

	if err == nil {
		t.Errorf("Expected error deleting reading by ID")
	}
}

func TestGetReadingsByDeviceId(t *testing.T) {
	reset()
	dbClient = newMockDb()

	expectedReadings, expectedNil := getReadingsByDeviceId(math.MaxInt32, "valid", "Pressure")

	if expectedReadings == nil {
		t.Errorf("Should return Readings")
	}

	if expectedNil != nil {
		t.Errorf("Should not throw error")
	}
}

func TestGetReadingsByDeviceIdLimited(t *testing.T) {
	reset()
	dbClient = newMockDb()

	for limit:= 0; limit < 5; limit++ {
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
	dbClient = newMockDb()

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
	// this uses memdb on purpose

	_, err := countReadings()

	if err != nil {
		t.Errorf("Unexpected error in CountReadings")
	}
}

func TestCountReadingsError(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := countReadings()

	if err == nil {
		t.Errorf("Expected error in CountReadings")
	}
}

func TestGetReadingsByDevice(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getReadingsByDevice("valid", 0)

	if err != nil {
		t.Errorf("Unexpected error in getReadingsByDevice")
	}
}

func TestGetReadingsByDeviceError(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getReadingsByDevice("error", 0)

	if err == nil {
		t.Errorf("Expected error in getReadingsByDevice")
	}
}

func TestGetReadingsByValueDescriptor(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getReadingsByValueDescriptor("valid", 0)

	if err != nil {
		t.Errorf("Unexpected error getting readings by value descriptor")
	}
}

func TestGetReadingsByValueDescriptorOverLimit(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getReadingsByValueDescriptor("", math.MaxInt32)

	if err == nil {
		t.Errorf("Expected error getting readings by value descriptor")
	}
}

func TestGetReadingsByValueDescriptorError(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getReadingsByValueDescriptor("error", 0)

	if err == nil {
		t.Errorf("Expected error in getting readings by value descriptor")
	}
}

func TestGetReadingsByValueDescriptorNames(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getReadingsByValueDescriptorNames([]string{"valid"}, 0)

	if err != nil {
		t.Errorf("Unexpected error getting readings by value descriptor names")
	}
}

func TestGetReadingsByValueDescriptorNamesError(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getReadingsByValueDescriptorNames([]string{"error"}, 0)

	if err == nil {
		t.Errorf("Expected error in getting readings by value descriptor names")
	}
}

func TestGetReadingsByCreationTime(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getReadingsByCreationTime(0xBEEF, 0 ,0)

	if err != nil {
		t.Errorf("Unexpected error getting readings by creation time")
	}
}

func TestGetReadingsByCreationTimeError(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getReadingsByCreationTime(0xDEADBEEF, 0, 0)

	if err == nil {
		t.Errorf("Expected error in getting readings by creation time")
	}
}

func TestGetReadingsByDeviceAndValueDescriptor(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getReadingsByDeviceAndValueDescriptor("valid", "valid", 0)

	if err != nil {
		t.Errorf("Unexpected error getting readings by device and value descriptor")
	}
}

func TestGetReadingsByDeviceAndValueDescriptorError(t *testing.T) {
	reset()
	dbClient = newMockDb()

	_, err := getReadingsByDeviceAndValueDescriptor("error", "error", 0)

	if err == nil {
		t.Errorf("Expected error in getting readings by device and value descriptor")
	}
}