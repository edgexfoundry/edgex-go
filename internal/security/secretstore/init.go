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
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	bootstrapContainer "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/container"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"
)

type Bootstrap struct {
	insecureSkipVerify bool
	vaultInterval      int
}

func NewBootstrapHandler(insecureSkipVerify bool, vaultInterval int) *Bootstrap {
	return &Bootstrap{
		insecureSkipVerify: insecureSkipVerify,
		vaultInterval:      vaultInterval,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func (b *Bootstrap) Handler(
	ctx context.Context,
	wg *sync.WaitGroup,
	startupTimer startup.Timer,
	dic *di.Container) bool {

	configuration := container.ConfigurationFrom(dic.Get)
	loggingClient := bootstrapContainer.LoggingClientFrom(dic.Get)

	//step 1: boot up secretstore general steps same as other EdgeX microservice

	//step 2: initialize the communications
	req := NewRequester(b.insecureSkipVerify, configuration.SecretService.CaFilePath, loggingClient)
	if req == nil {
		os.Exit(1)
	}

	vaultScheme := configuration.SecretService.Scheme
	vaultHost := fmt.Sprintf("%s:%v", configuration.SecretService.Server, configuration.SecretService.Port)
	intervalDuration := time.Duration(b.vaultInterval) * time.Second
	vc := secretstoreclient.NewSecretStoreClient(loggingClient, req, vaultScheme, vaultHost)

	//step 3: initialize and unseal Vault
	path := configuration.SecretService.TokenFolderPath
	filename := configuration.SecretService.TokenFile
	absPath := filepath.Join(path, filename)
	for shouldContinue := true; shouldContinue; {
		// Anonymous function used to prevent file handles from accumulating
		func() {
			tokenFile, err := os.OpenFile(absPath, os.O_CREATE|os.O_RDWR, 0600)
			if err != nil {
				loggingClient.Error(fmt.Sprintf("unable to open token file at %s with error: %s", absPath, err.Error()))
				os.Exit(1)
			}
			defer tokenFile.Close()
			sCode, _ := vc.HealthCheck()

			switch sCode {
			case http.StatusOK:
				loggingClient.Info(fmt.Sprintf("vault is initialized and unsealed (status code: %d)", sCode))
				shouldContinue = false
			case http.StatusTooManyRequests:
				loggingClient.Error(fmt.Sprintf("vault is unsealed and in standby mode (Status Code: %d)", sCode))
				shouldContinue = false
			case http.StatusNotImplemented:
				loggingClient.Info(fmt.Sprintf("vault is not initialized (status code: %d). Starting initialisation and unseal phases", sCode))
				_, err := vc.Init(configuration.SecretService, tokenFile)
				if err == nil {
					tokenFile.Seek(0, 0) // Read starting at beginning
					_, err = vc.Unseal(configuration.SecretService, tokenFile)
					if err == nil {
						shouldContinue = false
					}
				}
			case http.StatusServiceUnavailable:
				loggingClient.Info(fmt.Sprintf("vault is sealed (status code: %d). Starting unseal phase", sCode))
				_, err := vc.Unseal(configuration.SecretService, tokenFile)
				if err == nil {
					shouldContinue = false
				}
			default:
				if sCode == 0 {
					loggingClient.Error(fmt.Sprintf("vault is in an unknown state. No Status code available"))
				} else {
					loggingClient.Error(fmt.Sprintf("vault is in an unknown state. Status code: %d", sCode))
				}
			}
		}()

		if shouldContinue {
			loggingClient.Info(fmt.Sprintf("trying Vault init/unseal again in %d seconds", b.vaultInterval))
			time.Sleep(intervalDuration)
		}
	}

	/* After vault is init'd and unsealed, it takes a while to get ready to accept any request. During which period any request will get http 500 error.
	We need to check the status constantly until it return http StatusOK.
	*/
	ticker := time.NewTicker(time.Second)
	healthOkCh := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				if sCode, _ := vc.HealthCheck(); sCode == http.StatusOK {
					close(healthOkCh)
					ticker.Stop()
					return
				}
			}
		}
	}()

	// Wait on a StatusOK response from vc.HealthCheck()
	<-healthOkCh

	//Step 4: Launch token handler
	tokenProvider := NewTokenProvider(ctx, loggingClient)
	if configuration.SecretService.TokenProvider != "" {
		if err := tokenProvider.SetConfiguration(configuration.SecretService); err != nil {
			loggingClient.Error(fmt.Sprintf("failed to configure token provider: %s", err.Error()))
			os.Exit(1)
		}
		if err := tokenProvider.Launch(); err != nil {
			loggingClient.Error(fmt.Sprintf("token provider failed: %s", err.Error()))
			os.Exit(1)
		}
	} else {
		loggingClient.Info("no token provider configured")
	}

	// credential creation
	gk := NewGokeyGenerator(absPath)
	loggingClient.Warn("WARNING: The gokey generator is a reference implementation for credential generation and the underlying libraries not been reviewed for cryptographic security. The user is encouraged to perform their own security investigation before deployment.")
	cred := NewCred(req, absPath, gk, configuration.SecretService.GetSecretSvcBaseURL(), loggingClient)
	for dbname, info := range configuration.Databases {
		service := info.Service
		// generate credentials
		password, err := cred.GeneratePassword(dbname)
		if err != nil {
			loggingClient.Error(fmt.Sprintf("failed to generate credential pair for service %s", service))
			os.Exit(1)
		}
		pair := UserPasswordPair{
			User:     info.Username,
			Password: password,
		}

		// add credentials to service path if specified and they're not already there
		if len(service) != 0 {
			servicePath := fmt.Sprintf("/v1/secret/edgex/%s/mongodb", service)
			existing, err := cred.AlreadyInStore(servicePath)
			if err != nil {
				loggingClient.Error(err.Error())
				os.Exit(1)
			}
			if !existing {
				err = cred.UploadToStore(&pair, servicePath)
				if err != nil {
					loggingClient.Error(fmt.Sprintf("failed to upload credential pair for db %s on path %s", dbname, servicePath))
					os.Exit(1)
				}
			} else {
				loggingClient.Info(fmt.Sprintf("credentials for %s already present at path %s", dbname, servicePath))
			}
		}

		mongoPath := fmt.Sprintf("/v1/secret/edgex/mongo/%s", dbname)
		// add credentials to mongo path if they're not already there
		existing, err := cred.AlreadyInStore(mongoPath)
		if err != nil {
			loggingClient.Error(err.Error())
			os.Exit(1)
		}
		if !existing {
			err = cred.UploadToStore(&pair, mongoPath)
			if err != nil {
				loggingClient.Error(fmt.Sprintf("failed to upload credential pair for db %s on path %s", dbname, mongoPath))
				os.Exit(1)
			}
		} else {
			loggingClient.Info(fmt.Sprintf("credentials for %s already present at path %s", dbname, mongoPath))
		}
	}

	cert := NewCerts(req, configuration.SecretService.CertPath, absPath, configuration.SecretService.GetSecretSvcBaseURL(), loggingClient)
	existing, err := cert.AlreadyinStore()
	if err != nil {
		loggingClient.Error(err.Error())
		os.Exit(1)
	}

	if existing == true {
		loggingClient.Info("proxy certificate pair are in the secret store already, skip uploading")
		return false
	}

	loggingClient.Info("proxy certificate pair are not in the secret store yet, uploading them")
	cp, err := cert.ReadFrom(configuration.SecretService.CertFilePath, configuration.SecretService.KeyFilePath)
	if err != nil {
		loggingClient.Error("failed to get certificate pair from volume")
		os.Exit(1)
	}

	loggingClient.Info("proxy certificate pair are loaded from volume successfully, will upload to secret store")

	err = cert.UploadToStore(cp)
	if err != nil {
		loggingClient.Error("failed to upload the proxy cert pair into the secret store")
		loggingClient.Error(err.Error())
		os.Exit(1)
	}

	loggingClient.Info("proxy certificate pair are uploaded to secret store successfully, Vault init done successfully")
	return false
}
