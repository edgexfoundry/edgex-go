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
	"bufio"
	"context"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/messagebus/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

const (
	// file read/write permission only for owner
	readWriteOnlyForOwner os.FileMode = 0600
)

// Handler is the redis bootstrapping handler
type Handler struct {
	credentials bootstrapConfig.Credentials
}

// NewHandler instantiates a new Handler
func NewHandler() *Handler {
	return &Handler{}
}

// GetCredentials retrieves the message bus credentials from secretstore
func (handler *Handler) GetCredentials(ctx context.Context, _ *sync.WaitGroup, startupTimer startup.Timer,
	dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)

	var credentials = bootstrapConfig.Credentials{
		Username: "unset",
		Password: "unset",
	}

	for startupTimer.HasNotElapsed() {
		// retrieve message bus credentials from secretstore
		secrets, err := secretProvider.GetSecret(config.MessageQueue.SecretName)
		if err == nil {
			credentials.Username = secrets[secret.UsernameKey]
			credentials.Password = secrets[secret.PasswordKey]
			break
		}

		lc.Warnf("Could not retrieve message bus credentials (startup timer has not expired): %s", err.Error())
		startupTimer.SleepForInterval()
	}

	if credentials.Password == "unset" {
		lc.Error("Failed to retrieve message bus credentials before startup timer expired")
		return false
	}

	handler.credentials = credentials
	return true
}

// SetupPasswordFile sets up the message bus password file for MQTT specific
func (handler *Handler) SetupPasswordFile(ctx context.Context, _ *sync.WaitGroup, startupTimer startup.Timer,
	dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

	if config.MessageQueue.Type != "mqtt" {
		lc.Infof("messaage bus not mqtt type: %s, skip setting up password file", config.MessageQueue.Type)
		return true
	}

	pwdFile := config.MessageQueue.Optional["PasswordFile"]
	if len(pwdFile) == 0 {
		lc.Errorf("missing PasswordFile configuration for mqtt")
		return false
	}

	pwdFileBaseDir := filepath.Dir(pwdFile)
	if err := os.Chdir(pwdFileBaseDir); err != nil {
		lc.Errorf("failed to change directory %s: %v", pwdFileBaseDir, err)
		return false
	}

	// use mqtt password utility to set up password file
	cmd := exec.Command("mosquitto_passwd", "-c", "-b", pwdFile, handler.credentials.Username, handler.credentials.Password)
	_, err := cmd.Output()

	if err != nil {
		lc.Errorf("failed to execute command mosquitto_passwd: %v", err)
		return false
	}

	return true
}

// SetupConfFile dynamically creates message bus config file
func (handler *Handler) SetupConfFile(ctx context.Context, _ *sync.WaitGroup, _ startup.Timer,
	dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

	if config.MessageQueue.Type != "mqtt" {
		lc.Infof("message bus not mqtt type: %s, skip setting up config file", config.MessageQueue.Type)
		return true
	}

	brokerConfigFile := config.MessageQueue.Optional["BrokerConfigFile"]
	if len(brokerConfigFile) == 0 {
		lc.Errorf("missing brokerConfigFile configuration for mqtt")
		return false
	}

	configFileTemplate := `listener {{.MQTTPort}}
password_file {{.PwdFilePath}}`

	type mqttConfig struct {
		MQTTPort    int
		PwdFilePath string
	}

	mqttConf, err := template.New("mqtt-config").Parse(configFileTemplate + fmt.Sprintln())
	if err != nil {
		lc.Errorf("failed to parse MQTT config file template %s: %v", configFileTemplate, err)
		return false
	}

	lc.Infof("Creating the broker config file: %s", brokerConfigFile)

	// open config file with read-write and overwritten attribute (TRUNC)
	confFile, err := os.OpenFile(brokerConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, readWriteOnlyForOwner)
	if err != nil {
		lc.Errorf("failed to open config file %s: %v", brokerConfigFile, err)
		return false
	}
	defer func() {
		_ = confFile.Close()
	}()

	fwriter := bufio.NewWriter(confFile)
	if err := mqttConf.Execute(fwriter, mqttConfig{
		MQTTPort:    config.MessageQueue.Port,
		PwdFilePath: config.MessageQueue.Optional["PasswordFile"],
	}); err != nil {
		lc.Errorf("failed to execute mqttConfig template %s: %v", configFileTemplate, err)
		return false
	}

	if err := fwriter.Flush(); err != nil {
		lc.Errorf("failed to flush the mqtt config file writer buffer %v", err)
		return false
	}

	lc.Info("MQTT config file has been successfully set up")

	return true
}
