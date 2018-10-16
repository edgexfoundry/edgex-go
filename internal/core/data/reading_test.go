package data

import (
	"math"
	"testing"
)

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