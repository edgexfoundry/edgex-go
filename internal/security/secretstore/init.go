/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package secretstore

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	bootstrap "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-registry/registry"
)

// Global variables
var Configuration = &ConfigurationStruct{}
var LoggingClient logger.LoggingClient

type Handler struct {
	insecureSkipVerify bool
}

func NewHandler(insecureSkipVerify bool) Handler {
	return Handler{
		insecureSkipVerify: insecureSkipVerify,
	}
}

func (h Handler) BootstrapHandler(
	wg *sync.WaitGroup,
	ctx context.Context,
	startupTimer startup.Timer,
	config bootstrap.Configuration,
	logging logger.LoggingClient,
	registry registry.Client) bool {

	// update global variables.
	LoggingClient = logging

	//step 1: boot up secretstore general steps same as other EdgeX microservice

	//step 2: initialize the communications
	req := NewRequester(h.insecureSkipVerify)

	//step 3: initialize and unseal Vault

	//Step 4:
	//TODO: create vault access token for different roles

	//step 5 :
	//TODO: implement credential creation

	absTokenPath := filepath.Join(Configuration.SecretService.TokenFolderPath, Configuration.SecretService.TokenFile)
	cert := NewCerts(req, Configuration.SecretService.CertPath, absTokenPath)
	existing, err := cert.AlreadyinStore()
	if err != nil {
		LoggingClient.Error(err.Error())
		return false
	}

	if existing == true {
		LoggingClient.Info("proxy certificate pair are in the secret store already, skip uploading")
		os.Exit(0)
	}

	LoggingClient.Info("proxy certificate pair are not in the secret store yet, uploading them")
	cp, err := cert.ReadFrom(Configuration.SecretService.CertFilePath, Configuration.SecretService.KeyFilePath)
	if err != nil {
		LoggingClient.Error("failed to get certificate pair from volume")
		return false
	}

	LoggingClient.Info("proxy certificate pair are loaded from volume successfully, will upload to secret store")

	err = cert.UploadToStore(cp)
	if err != nil {
		LoggingClient.Error("failed to upload the proxy cert pair into the secret store")
		LoggingClient.Error(err.Error())
		return false
	}

	LoggingClient.Info("proxy certificate pair are uploaded to secret store successfully, Vault init done successfully")
	return false
}
