/*******************************************************************************
 * Copyright (C) 2025 IOTech Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package createtoken

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/config"
	"github.com/edgexfoundry/edgex-go/internal/security/fileprovider/container"
	secretStoreConfig "github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"

	"github.com/stretchr/testify/require"
)

const mockEntityId = "mockId123"

func mockDic() *di.Container {

	return di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{
				LogLevel:          "",
				SecretStore:       secretStoreConfig.SecretStoreInfo{},
				TokenFileProvider: config.TokenFileProviderInfo{},
			}
		},
	})
}

func TestNewCommand(t *testing.T) {
	// Arrange
	dic := mockDic()

	tests := []struct {
		name        string
		cmdArgs     []string
		expectedErr bool
	}{
		{"Good: createToken required --entityId option", []string{"--entityId=" + mockEntityId}, false},
		{"Bad: createToken empty option", []string{""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := NewCommand(dic, tt.cmdArgs)
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, command)
			}
		})
	}
}
