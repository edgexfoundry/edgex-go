package messaging

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
	"github.com/stretchr/testify/assert"

	"github.com/edgexfoundry/edgex-go/internal/core/data/config"
	"github.com/edgexfoundry/edgex-go/internal/core/data/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	messaging2 "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

var lc logger.LoggingClient
var dic *di.Container
var usernameSecretData = map[string]string{
	messaging2.SecretUsernameKey: "username",
	messaging2.SecretPasswordKey: "password",
}

func TestMain(m *testing.M) {
	lc = logger.NewMockClient()

	dic = di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
	})

	os.Exit(m.Run())
}

func TestBootstrapHandler(t *testing.T) {
	validCreateClient := config.ConfigurationStruct{
		MessageQueue: bootstrapConfig.MessageBusInfo{
			Type:               messaging.Redis,
			Protocol:           "redis",
			Host:               "localhost",
			Port:               6379,
			PublishTopicPrefix: "edgex/events/#",
			AuthMode:           messaging2.AuthModeUsernamePassword,
			SecretName:         "redisdb",
		},
	}

	invalidSecrets := config.ConfigurationStruct{
		MessageQueue: bootstrapConfig.MessageBusInfo{
			AuthMode:   messaging2.AuthModeCert,
			SecretName: "redisdb",
		},
	}

	invalidNoConnect := config.ConfigurationStruct{
		MessageQueue: bootstrapConfig.MessageBusInfo{
			Type:       messaging.MQTT, // This will cause no connection since broker not available
			Protocol:   "tcp",
			Host:       "localhost",
			Port:       8765,
			AuthMode:   messaging2.AuthModeUsernamePassword,
			SecretName: "redisdb",
		},
	}

	tests := []struct {
		Name           string
		Config         *config.ConfigurationStruct
		ExpectedResult bool
		ExpectClient   bool
	}{
		{"Valid - creates client", &validCreateClient, true, true},
		{"Invalid - secrets error", &invalidSecrets, false, false},
		{"Invalid - can't connect", &invalidNoConnect, false, false},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			provider := &mocks.SecretProvider{}
			provider.On("GetSecret", test.Config.MessageQueue.SecretName).Return(usernameSecretData, nil)
			dic.Update(di.ServiceConstructorMap{
				container.ConfigurationName: func(get di.Get) interface{} {
					return test.Config
				},
				bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
					return provider
				},
				bootstrapContainer.MessagingClientName: func(get di.Get) interface{} {
					return nil
				},
			})

			actual := BootstrapHandler(context.Background(), &sync.WaitGroup{}, startup.NewTimer(1, 1), dic)
			assert.Equal(t, test.ExpectedResult, actual)
			assert.Empty(t, test.Config.MessageQueue.Optional)
			if test.ExpectClient {
				assert.NotNil(t, bootstrapContainer.MessagingClientFrom(dic.Get))
			} else {
				assert.Nil(t, bootstrapContainer.MessagingClientFrom(dic.Get))
			}
		})
	}
}
