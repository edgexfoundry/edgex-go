package data

import (
	"math"
	"testing"
)

func TestCheckMaxLimit(t *testing.T) {
	reset()

	testedLimit := math.MinInt32

	expectedNil := checkMaxLimit(testedLimit)

	if expectedNil != nil {
		t.Errorf("Should not exceed limit")
	}
}

func TestCheckMaxLimitOverLimit(t *testing.T) {
	reset()

	testedLimit := math.MaxInt32

	expectedErr := checkMaxLimit(testedLimit)

	if expectedErr == nil {
		t.Errorf("Exceeded limit and should throw error")
	}
}