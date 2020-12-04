package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateKey(t *testing.T) {
	result := CreateKey(EventsCollectionDeviceName, "TestDeviceName")
	expected := EventsCollectionDeviceName + DBKeySeparator + "TestDeviceName"
	assert.Equal(t, expected, result)
}
