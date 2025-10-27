//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"errors"
	"os"
	"path"
	"sync"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/postgres/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/postgres/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSetupDBScriptFiles(t *testing.T) {
	secProviderExt := &mocks.SecretProviderExt{}
	secProviderExt.On("ListSecretNames", mock.Anything).Return([]string{}, nil)

	mockDic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				DatabaseConfig: config.DatabaseBootstrapInitInfo{
					Path: "",
					Name: "",
				},
			}
		},
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		bootstrapContainer.SecretProviderExtName: func(get di.Get) interface{} {
			return secProviderExt
		},
	})
	result := SetupDBScriptFiles(context.Background(), &sync.WaitGroup{}, startup.NewStartUpTimer("mockSvcKey"), mockDic)
	require.False(t, result)

	mockDir := "testDir"
	mockFileName := "testFilename"
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(mockDir)

	mockDic.Update(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				DatabaseConfig: config.DatabaseBootstrapInitInfo{
					Path: mockDir,
					Name: mockFileName,
				},
			}
		}})
	result = SetupDBScriptFiles(context.Background(), &sync.WaitGroup{}, startup.NewStartUpTimer("mockSvcKey"), mockDic)
	require.True(t, result)
}

func TestGetServiceCredentialsInvalidSecret(t *testing.T) {
	invalidSecretPath := "invalidPath"
	secProviderExt := &mocks.SecretProviderExt{}
	secProviderExt.On("ListSecretNames", mock.Anything).Return([]string{invalidSecretPath}, nil)
	secProviderExt.On("HasSecret", path.Join(invalidSecretPath, postgresSecretName)).Return(false, errors.New("secret not found"))

	mockDic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		bootstrapContainer.SecretProviderExtName: func(get di.Get) interface{} {
			return secProviderExt
		},
	})

	fileName := "testScriptFile"
	testScriptFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	require.NoError(t, err)
	defer func() {
		_ = testScriptFile.Close()
		_ = os.RemoveAll(fileName)
	}()

	err = getServiceCredentials(mockDic, testScriptFile)
	require.Error(t, err)
}

func TestGetServiceCredentialsValid(t *testing.T) {
	validSecretPath := "validPath"
	secProviderExt := &mocks.SecretProviderExt{}
	expectedUsername := "core-data"
	expectedPassword := "password1234"

	expectedResult := map[string]string{"username": expectedUsername, "password": expectedPassword}
	secProviderExt.On("ListSecretNames", mock.Anything).Return([]string{validSecretPath}, nil)
	secProviderExt.On("HasSecret", path.Join(validSecretPath, postgresSecretName)).Return(true, nil)
	secProviderExt.On("GetSecret", path.Join(validSecretPath, postgresSecretName)).Return(expectedResult, nil)

	mockDic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		bootstrapContainer.SecretProviderExtName: func(get di.Get) interface{} {
			return secProviderExt
		},
	})

	fileName := "testScriptFile"
	testScriptFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	require.NoError(t, err)
	defer func() {
		_ = testScriptFile.Close()
		_ = os.RemoveAll(fileName)
	}()

	err = getServiceCredentials(mockDic, testScriptFile)
	require.NoError(t, err)
}
