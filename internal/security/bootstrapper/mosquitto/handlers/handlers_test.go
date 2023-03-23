/*******************************************************************************
* Copyright 2022 Intel Corporation
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

package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/mosquitto/config"
	messagebus "github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/mosquitto/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

type Testdata struct {
	lc         logger.LoggingClient
	dic        *di.Container
	secretData map[string]string
	ctx        context.Context
}
type args struct {
	ctx          context.Context
	startupTimer startup.Timer
	dic          *di.Container
}

func setUp(t *testing.T, secretName string, brokerConfigFile string, passwordFile string) *Testdata {

	mockLc := logger.NewMockClient()
	configuration := &config.ConfigurationStruct{
		SecureMosquitto: config.SecureMosquittoInfo{
			SecretName:       secretName,
			BrokerConfigFile: brokerConfigFile,
			PasswordFile:     passwordFile,
		},
	}

	return &Testdata{
		lc: mockLc,
		dic: di.NewContainer(di.ServiceConstructorMap{
			container.LoggingClientInterfaceName: func(get di.Get) interface{} {
				return mockLc
			},
			messagebus.ConfigurationName: func(get di.Get) interface{} {
				return configuration
			},
		}),
		secretData: map[string]string{
			messaging.SecretUsernameKey: "username",
			messaging.SecretPasswordKey: "password",
		},
		ctx: context.Background(),
	}
}
func TestHandler_GetCredentials(t *testing.T) {

	// setup mock secret client
	expectedSecretData := map[string]string{
		"username": "TEST_USER",
		"password": "TEST_PASS",
	}

	mockSecretProvider := &mocks.SecretProvider{}
	mockSecretProvider.On("GetSecret", "message-bus").Return(expectedSecretData, nil)

	mockSecretProvider.On("GetSecret", "").Return(nil, errors.New("Empty Secret Name"))
	mockSecretProvider.On("GetSecret", "notfound").Return(nil, errors.New("Not Found"))

	tests := []struct {
		name       string
		args       args
		secretName string
		want       bool
	}{
		{
			name:       "GetCredentials ok",
			secretName: "message-bus",
			want:       true,
		},
		{
			name:       "GetCredentials no secret name",
			secretName: "",
			want:       false,
		},
		{
			name:       "GetCredentials secret name not found",
			secretName: "notfound",
			want:       false,
		},
	}

	for _, tt := range tests {
		testData := setUp(t, tt.secretName, "", "")
		args := args{
			startupTimer: startup.NewTimer(3, 1),
			ctx:          testData.ctx,
			dic:          testData.dic,
		}
		testData.dic.Update(di.ServiceConstructorMap{
			container.SecretProviderName: func(get di.Get) interface{} {
				return mockSecretProvider
			},
		})

		t.Run(tt.name, func(t *testing.T) {
			handler := &Handler{}
			if got := handler.GetCredentials(args.ctx, nil, args.startupTimer, args.dic); got != tt.want {
				t.Errorf("Handler.GetCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestHandler_SetupConfFile(t *testing.T) {
	tests := []struct {
		name             string
		args             args
		brokerConfigFile string
		msqQueuePort     int
		want             bool
	}{
		{
			name:             "SetupConfFile ok",
			brokerConfigFile: "/tmp/mosquitto.conf",
			want:             true,
		},
		{
			name:             "SetupConfFile broker file name not valid",
			brokerConfigFile: "/bad/file/path",
			want:             false,
		},
		{
			name:             "SetupConfFile config file name not set",
			brokerConfigFile: "",
			want:             false,
		},
	}

	for _, tt := range tests {
		testData := setUp(t, "", tt.brokerConfigFile, "")
		args := args{
			startupTimer: startup.NewTimer(3, 1),
			ctx:          testData.ctx,
			dic:          testData.dic,
		}
		t.Run(tt.name, func(t *testing.T) {
			handler := &Handler{}
			if got := handler.SetupMosquittoConfFile(args.ctx, nil, args.startupTimer, args.dic); got != tt.want {
				t.Errorf("Handler.GetCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}
