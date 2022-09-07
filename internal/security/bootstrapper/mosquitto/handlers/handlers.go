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

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/mosquitto/container"

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

// GetCredentials retrieves the mosquitto credentials from secretstore
func (handler *Handler) GetCredentials(ctx context.Context, _ *sync.WaitGroup, startupTimer startup.Timer,
	dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)

	var credentials *bootstrapConfig.Credentials

	for startupTimer.HasNotElapsed() {
		// retrieve mosquitto credentials from secretstore
		secrets, err := secretProvider.GetSecret(config.SecureMosquitto.SecretName)
		if err == nil {
			credentials = &bootstrapConfig.Credentials{
				Username: secrets[secret.UsernameKey],
				Password: secrets[secret.PasswordKey],
			}
			break
		}

		lc.Warnf("Could not retrieve mosquitto credentials (startup timer has not expired): %s", err.Error())
		startupTimer.SleepForInterval()
	}

	if credentials == nil {
		lc.Error("Failed to retrieve mosquitto credentials before startup timer expired")
		return false
	}

	handler.credentials = *credentials
	return true
}

// SetupMosquittoPasswordFile sets up the mosquitto password file
func (handler *Handler) SetupMosquittoPasswordFile(ctx context.Context, _ *sync.WaitGroup, startupTimer startup.Timer,
	dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

	pwdFile := config.SecureMosquitto.PasswordFile
	if len(pwdFile) == 0 {
		lc.Errorf("missing PasswordFile configuration for mqtt")
		return false
	}

	pwdFileBaseDir := filepath.Dir(pwdFile)
	if err := os.Chdir(pwdFileBaseDir); err != nil {
		lc.Errorf("failed to change directory %s: %v", pwdFileBaseDir, err)
		return false
	}

	cmd := exec.Command("mosquitto_passwd", "-c", "-b", pwdFile, handler.credentials.Username, handler.credentials.Password)
	_, err := cmd.Output()

	if err != nil {
		lc.Errorf("failed to execute command mosquitto_passwd: %v", err)
		return false
	}

	return true
}

// SetupMosquittoConfFile dynamically creates mosquitto config file
func (handler *Handler) SetupMosquittoConfFile(ctx context.Context, _ *sync.WaitGroup, _ startup.Timer,
	dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

	brokerConfigFile := config.SecureMosquitto.BrokerConfigFile
	if len(brokerConfigFile) == 0 {
		lc.Errorf("missing brokerConfigFile configuration for mosquitto")
		return false
	}

	configFileTemplate := `listener {{.MQTTPort}}
password_file {{.PwdFilePath}}`

	type mosquittoConfig struct {
		MQTTPort    int
		PwdFilePath string
	}

	mosquittoConf, err := template.New("mosquitto-config").Parse(configFileTemplate + fmt.Sprintln())
	if err != nil {
		lc.Errorf("failed to parse mosquitto config file template %s: %v", configFileTemplate, err)
		return false
	}

	lc.Infof("Creating the mosquitto config file: %s", brokerConfigFile)

	// open config file with read-write and overwritten attribute (TRUNC)
	confFile, err := os.OpenFile(brokerConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, readWriteOnlyForOwner)
	if err != nil {
		lc.Errorf("failed to open mosquitto config file %s: %v", brokerConfigFile, err)
		return false
	}
	defer func() {
		_ = confFile.Close()
	}()

	fwriter := bufio.NewWriter(confFile)
	if err := mosquittoConf.Execute(fwriter, mosquittoConfig{
		MQTTPort:    config.SecureMosquitto.Port,
		PwdFilePath: config.SecureMosquitto.PasswordFile,
	}); err != nil {
		lc.Errorf("failed to execute mosquittoConfig template %s: %v", configFileTemplate, err)
		return false
	}

	if err := fwriter.Flush(); err != nil {
		lc.Errorf("failed to flush the mqtt config file writer buffer %v", err)
		return false
	}

	lc.Info("mosquitto config file has been successfully set up")

	return true
}
