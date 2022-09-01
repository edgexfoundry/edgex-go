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

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/messagebus/config"
	messagebus "github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/messagebus/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

type Testdata struct {
	lc         logger.LoggingClient
	dic        *di.Container
	secretData map[string]string
	ctx        context.Context
}

func setUp(t *testing.T, secretName string, brokerConfigFile string,
	msqQueueType string, passwordFile string) *Testdata {

	mockLc := logger.NewMockClient()
	configuration := &config.ConfigurationStruct{
		MessageQueue: config.MessageQueueInfo{
			SecretName:       secretName,
			Type:             msqQueueType,
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
	mockSecretProvider.On("GetSecret", "mqtt-bus").Return(expectedSecretData, nil)

	mockSecretProvider.On("GetSecret", "").Return(nil, errors.New("Empty Secret Name"))
	mockSecretProvider.On("GetSecret", "notfound").Return(nil, errors.New("Not Found"))

	type args struct {
		ctx          context.Context
		startupTimer startup.Timer
		dic          *di.Container
	}
	tests := []struct {
		name       string
		args       args
		secretName string
		want       bool
	}{
		{
			name: "GetCredentials ok",
			args: args{

				startupTimer: startup.NewTimer(3, 1),
			},
			secretName: "mqtt-bus",
			want:       true,
		},
		{
			name: "GetCredentials no secret name",
			args: args{

				startupTimer: startup.NewTimer(3, 1),
			},
			secretName: "",
			want:       false,
		},
		{
			name: "GetCredentials secret name not found",
			args: args{

				startupTimer: startup.NewTimer(3, 1),
			},
			secretName: "notfound",
			want:       false,
		},
	}

	for _, tt := range tests {
		testData := setUp(t, tt.secretName, "", "", "")
		tt.args.ctx = testData.ctx
		tt.args.dic = testData.dic
		testData.dic.Update(di.ServiceConstructorMap{
			container.SecretProviderName: func(get di.Get) interface{} {
				return mockSecretProvider
			},
		})

		t.Run(tt.name, func(t *testing.T) {
			handler := &Handler{}
			if got := handler.GetCredentials(tt.args.ctx, nil, tt.args.startupTimer, tt.args.dic); got != tt.want {
				t.Errorf("Handler.GetCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestHandler_SetupConfFile(t *testing.T) {
	type args struct {
		ctx          context.Context
		startupTimer startup.Timer
		dic          *di.Container
	}
	tests := []struct {
		name             string
		args             args
		brokerConfigFile string
		msgQueueType     string
		msqQueuePort     int
		want             bool
	}{
		{
			name: "SetupConfFile ok",
			args: args{

				startupTimer: startup.NewTimer(3, 1),
			},
			brokerConfigFile: "/tmp/mosquitto.conf",
			msgQueueType:     "mqtt",
			want:             true,
		},
		{
			name: "SetupConfFile No message queue type set",
			args: args{

				startupTimer: startup.NewTimer(3, 1),
			},
			brokerConfigFile: "/tmp/mosquitto.conf",
			msgQueueType:     "",
			want:             true,
		},
		{
			name: "SetupConfFile broker file name not valid",
			args: args{

				startupTimer: startup.NewTimer(3, 1),
			},
			brokerConfigFile: "/bad/file/path",
			msgQueueType:     "mqtt",
			want:             false,
		},
		{
			name: "SetupConfFile config file name not set",
			args: args{

				startupTimer: startup.NewTimer(3, 1),
			},
			brokerConfigFile: "",
			msgQueueType:     "mqtt",
			want:             false,
		},
	}

	for _, tt := range tests {
		testData := setUp(t, "", tt.brokerConfigFile, tt.msgQueueType, "")
		tt.args.ctx = testData.ctx
		tt.args.dic = testData.dic

		t.Run(tt.name, func(t *testing.T) {
			handler := &Handler{}
			if got := handler.SetupConfFile(tt.args.ctx, nil, tt.args.startupTimer, tt.args.dic); got != tt.want {
				t.Errorf("Handler.GetCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestHandler_SetupPasswordFile(t *testing.T) {
	type args struct {
		ctx          context.Context
		startupTimer startup.Timer
		dic          *di.Container
	}
	tests := []struct {
		name         string
		args         args
		passwordFile string
		msgQueueType string
		msqQueuePort int
		want         bool
	}{
		{
			name: "SetupPasswordFile no mosquitto",
			args: args{

				startupTimer: startup.NewTimer(3, 1),
			},
			passwordFile: "/tmp/passwd",
			msgQueueType: "mqtt",
			want:         false,
		},
		{
			name: "SetupPasswordFile No message queue type set",
			args: args{

				startupTimer: startup.NewTimer(3, 1),
			},
			passwordFile: "/tmp/passwd",
			msgQueueType: "",
			want:         true,
		},
		{
			name: "SetupPasswordFile password file name not valid",
			args: args{

				startupTimer: startup.NewTimer(3, 1),
			},
			passwordFile: "/bad/file/path",
			msgQueueType: "mqtt",
			want:         false,
		},
		{
			name: "SetupPasswordFile password file name not set",
			args: args{

				startupTimer: startup.NewTimer(3, 1),
			},
			passwordFile: "",
			msgQueueType: "mqtt",
			want:         false,
		},
	}

	for _, tt := range tests {
		testData := setUp(t, "", "", tt.msgQueueType, tt.passwordFile)
		tt.args.ctx = testData.ctx
		tt.args.dic = testData.dic

		t.Run(tt.name, func(t *testing.T) {
			handler := &Handler{}
			if got := handler.SetupPasswordFile(tt.args.ctx, nil, tt.args.startupTimer, tt.args.dic); got != tt.want {
				t.Errorf("Handler.GetCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}
