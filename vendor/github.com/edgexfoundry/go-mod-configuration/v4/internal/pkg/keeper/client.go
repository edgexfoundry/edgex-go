//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/spf13/cast"

	httpClient "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/http"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-messaging/v4/messaging"
	msgTypes "github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

	"github.com/edgexfoundry/go-mod-configuration/v4/pkg/types"
)

const (
	keeperTopicPrefix = "edgex/configs"
	KeyDelimiter      = "/"
)

type keeperClient struct {
	keeperUrl      string
	configBasePath string
	watchingDone   chan bool

	commonClient interfaces.CommonClient
	kvsClient    interfaces.KVSClient
}

// NewKeeperClient creates a new Keeper Client.
func NewKeeperClient(config types.ServiceConfig) *keeperClient {
	client := keeperClient{
		keeperUrl:      config.GetUrl(),
		configBasePath: config.BasePath,
		watchingDone:   make(chan bool, 1),
	}

	// Create the common and KVS http clients for invoking APIs from Keeper
	client.commonClient = httpClient.NewCommonClient(client.keeperUrl, config.AuthInjector)
	client.kvsClient = httpClient.NewKVSClient(client.keeperUrl, config.AuthInjector)
	return &client
}

func (k *keeperClient) fullPath(name string) string {
	return path.Join(k.configBasePath, name)
}

// IsAlive simply checks if Core Keeper is up and running at the configured URL
func (k *keeperClient) IsAlive() bool {
	if _, err := k.commonClient.Ping(context.Background()); err != nil {
		return false
	}
	return true
}

// HasConfiguration checks to see if Core Keeper contains the service's configuration.
func (k *keeperClient) HasConfiguration() (bool, error) {
	_, err := k.kvsClient.ListKeys(context.Background(), k.configBasePath)
	if err != nil {
		if err.Code() == http.StatusNotFound {
			return false, nil
		}
		return false, fmt.Errorf("checking configuration existence from Core Keeper failed: %v", err)
	}
	return true, nil
}

// HasSubConfiguration checks to see if the Configuration service contains the service's sub configuration.
func (k *keeperClient) HasSubConfiguration(name string) (bool, error) {
	keyPath := k.fullPath(name)
	_, err := k.kvsClient.ListKeys(context.Background(), keyPath)
	if err != nil {
		if err.Code() == http.StatusNotFound {
			return false, nil
		}
		return false, fmt.Errorf("checking sub configuration existence from Core Keeper failed: %v", err)
	}
	return true, nil
}

// PutConfigurationMap puts a full configuration map into Core Keeper.
// The sub-paths to where the values are to be stored in Core Keeper are generated from the map key.
func (k *keeperClient) PutConfigurationMap(configuration map[string]any, overwrite bool) error {
	keyValues := convertInterfaceToPairs("", configuration)

	// Put config properties into Core Keeper.
	for _, keyValue := range keyValues {
		exists, _ := k.ConfigurationValueExists(keyValue.Key)
		if !exists || overwrite {
			if err := k.PutConfigurationValue(keyValue.Key, []byte(keyValue.Value)); err != nil {
				return err
			}
		}
	}

	return nil
}

// PutConfiguration puts a full configuration struct into the Configuration provider
func (k *keeperClient) PutConfiguration(config interface{}, overwrite bool) error {
	var err error
	if overwrite {
		value := config
		if byteArray, ok := config.([]byte); ok {
			value = string(byteArray)
		}
		request := requests.UpdateKeysRequest{
			Value: value,
		}
		_, err = k.kvsClient.UpdateValuesByKey(context.Background(), k.configBasePath, true, request)
	} else {
		kvPairs := convertInterfaceToPairs("", config)
		for _, kv := range kvPairs {
			exists, err := k.ConfigurationValueExists(kv.Key)
			if err != nil {
				return err
			}
			if !exists {
				// Only create the key if not exists in core keeper
				if err = k.PutConfigurationValue(kv.Key, []byte(kv.Value)); err != nil {
					return err
				}
			}
		}
	}
	if err != nil {
		return fmt.Errorf("error occurred while creating/updating configuration, error: %v", err)
	}
	return nil
}

// GetConfiguration gets the full configuration from Core Keeper into the target configuration struct.
// Passed in struct is only a reference for decoder, empty struct is ok
// Returns the configuration in the target struct as interface{}, which caller must cast
func (k *keeperClient) GetConfiguration(configStruct interface{}) (interface{}, error) {
	exists, err := k.HasConfiguration()
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("the Configuration service (EdgeX Keeper) doesn't contain configuration for %s", k.configBasePath)
	}

	resp, err := k.kvsClient.ValuesByKey(context.Background(), k.configBasePath)
	if err != nil {
		return nil, err
	}

	err = decode(k.configBasePath+KeyDelimiter, resp.Response, configStruct)
	if err != nil {
		return nil, err
	}
	return configStruct, nil
}

func (k *keeperClient) WatchForChanges(updateChannel chan<- interface{}, errorChannel chan<- error, configuration interface{}, waitKey string, getMsgClientCb func() messaging.MessageClient) {
	messageClient := getMsgClientCb()
	if messageClient == nil {
		configErr := errors.New("unable to use MessageClient to watch for configuration changes")
		errorChannel <- configErr
		return
	}

	messages := make(chan msgTypes.MessageEnvelope)
	topic := path.Join(keeperTopicPrefix, k.configBasePath, waitKey, "#")
	topics := []msgTypes.TopicChannel{
		{
			Topic:    topic,
			Messages: messages,
		},
	}

	watchErrors := make(chan error)
	err := messageClient.Subscribe(topics, watchErrors)
	if err != nil {
		_ = messageClient.Disconnect()
		errorChannel <- err
		return
	}

	go func() {
		defer func() {
			_ = messageClient.Disconnect()
		}()

		// send a nil value to updateChannel once the watcher connection is established
		// for go-mod-bootstrap to ignore the first change event
		// refer to the isFirstUpdate variable declared in https://github.com/edgexfoundry/go-mod-bootstrap/blob/main/bootstrap/config/config.go
		updateChannel <- nil

	outerLoop:
		for {
			select {
			case <-k.watchingDone:
				return
			case e := <-watchErrors:
				errorChannel <- e
			case msgEnvelope := <-messages:
				if msgEnvelope.ContentType != common.ContentTypeJSON {
					errorChannel <- fmt.Errorf("invalid content type of configuration changes message, expected: %s, but got: %s", common.ContentTypeJSON, msgEnvelope.ContentType)
					continue
				}
				var updatedConfig models.KVS
				// unmarshal the updated config to KV DTO
				updatedConfig, err := msgTypes.GetMsgPayload[models.KVS](msgEnvelope)
				if err != nil {
					errorChannel <- fmt.Errorf("failed to unmarshal the updated configuration: %v", err)
					continue
				}
				keyPrefix := path.Join(k.configBasePath, waitKey)

				// get the whole configs KV DTO array from Keeper with the same keyPrefix
				kvConfigs, err := k.kvsClient.ValuesByKey(context.Background(), keyPrefix)
				if err != nil {
					errorChannel <- fmt.Errorf("failed to get the configurations with key prefix %s from Keeper: %v", keyPrefix, err)
					continue
				}

				// if the updated key not equal to keyPrefix, need to check the updated key and value from the message payload are valid
				// e.g. keyPrefix = "edgex/3.0/core-data/Writable" which is the root level of Writable configuration
				if updatedConfig.Key != keyPrefix {
					foundUpdatedKey := false
					for _, c := range kvConfigs.Response {
						if c.Key == updatedConfig.Key {
							// the updated key from the message payload has been found in Keeper
							foundUpdatedKey = true
							// if the updated value in the message payload is different from the one obtained by Keeper
							// skip this subscribed message payload and continue the outer loop
							if c.Value != updatedConfig.Value {
								continue outerLoop
							}
							break
						}
					}
					// if the updated key from the message payload hasn't been found in Keeper
					// skip this subscribed message payload
					if !foundUpdatedKey {
						errorChannel <- fmt.Errorf("the updated key %s hasn't been found in Keeper, skipping this message", updatedConfig.Key)
						continue
					}
				}

				// decode KV DTO array to configuration struct
				err = decode(keyPrefix, kvConfigs.Response, configuration)
				if err != nil {
					errorChannel <- fmt.Errorf("failed to decode the updated configuration: %v", err)
					continue
				}
				updateChannel <- configuration
			}
		}
	}()
}

// StopWatching causes all WatchForChanges processing to stop
func (k *keeperClient) StopWatching() {
	k.watchingDone <- true
}

// ConfigurationValueExists checks if a configuration value exists in Core Keeper
func (k *keeperClient) ConfigurationValueExists(name string) (bool, error) {
	keyPath := k.fullPath(name)
	_, err := k.kvsClient.ListKeys(context.Background(), keyPath)
	if err != nil {
		if err.Code() == http.StatusNotFound {
			return false, nil
		}
		return false, fmt.Errorf("checking configuration existence from Core Keeper failed: %v", err)
	}
	return true, nil
}

// GetConfigurationValue gets a specific configuration value from Core Keeper
func (k *keeperClient) GetConfigurationValue(name string) ([]byte, error) {
	keyPath := k.fullPath(name)
	return k.GetConfigurationValueByFullPath(keyPath)
}

// GetConfigurationValueByFullPath gets a specific configuration value given the full path from Core Keeper
func (k *keeperClient) GetConfigurationValueByFullPath(fullPath string) ([]byte, error) {
	resp, err := k.kvsClient.ValuesByKey(context.Background(), fullPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get value for %s from Core Keeper: %v", fullPath, err)
	}
	if len(resp.Response) == 0 {
		return nil, fmt.Errorf("%s configuration not found", fullPath)
	}

	valueStr := cast.ToString(resp.Response[0].Value)
	return []byte(valueStr), nil
}

// PutConfigurationValue puts a specific configuration value into Core Keeper
func (k *keeperClient) PutConfigurationValue(name string, value []byte) error {
	keyPath := k.fullPath(name)
	request := requests.UpdateKeysRequest{
		Value: string(value),
	}
	_, err := k.kvsClient.UpdateValuesByKey(context.Background(), keyPath, false, request)
	if err != nil {
		return fmt.Errorf("unable to put value for %s into Core Keeper: %v", keyPath, err)
	}
	return nil
}

// GetConfigurationKeys returns all keys under name
func (k *keeperClient) GetConfigurationKeys(name string) ([]string, error) {
	keyPath := k.fullPath(name)
	resp, err := k.kvsClient.ListKeys(context.Background(), keyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get list of keys for %s from Core Keeper: %v", keyPath, err)
	}

	var list []string
	for _, v := range resp.Response {
		list = append(list, string(v))
	}
	return list, nil
}
