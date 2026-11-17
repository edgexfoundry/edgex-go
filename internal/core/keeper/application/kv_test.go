//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/edgexfoundry/edgex-go/internal/core/keeper/container"
	dbMock "github.com/edgexfoundry/edgex-go/internal/core/keeper/infrastructure/interfaces/mocks"
)

func TestAddKeysFlattenValidation(t *testing.T) {
	dbClientMock := &dbMock.DBClient{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
	})

	// flatten with a field name containing a character rejected by ValidateKeys
	// must be refused before any database call, see issue #5495
	dbClientMock.On("AddKeeperKeys", mock.Anything, true).Return([]models.KeyOnly{"edgex/v4/test/Log:Level"}, nil).Maybe()

	keys, err := AddKeys(models.KVS{Key: "edgex/v4/test", Value: map[string]any{"Log:Level": "DEBUG"}}, true, dic)
	assert.Error(t, err)
	assert.Nil(t, keys)
	dbClientMock.AssertNotCalled(t, "AddKeeperKeys", mock.Anything, true)

	// flatten with valid field names must reach the database
	dbClientMock2 := &dbMock.DBClient{}
	dbClientMock2.On("AddKeeperKeys", mock.Anything, true).Return([]models.KeyOnly{"edgex/v4/test/MaxEventSize"}, nil)
	dic2 := di.NewContainer(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock2
		},
	})
	keys, err = AddKeys(models.KVS{Key: "edgex/v4/test", Value: map[string]any{"MaxEventSize": 25000}}, true, dic2)
	assert.NoError(t, err)
	assert.Equal(t, []models.KeyOnly{"edgex/v4/test/MaxEventSize"}, keys)
}
