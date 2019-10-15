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
 * @author: Alain Pulluelo, ForgeRock AS
 * @author: Tingyu Zeng, Dell
 * @author: Daniel Harms, Dell
 *
 *******************************************************************************/

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func main() {
	if len(os.Args) < 2 {
		usage.HelpCallbackSecuritySecretStore()
	}

	var initNeeded bool
	var insecureSkipVerify bool
	var configFileLocation string
	var vaultInterval int
	var configDir, profileDir string
	var useRegistry bool

	flag.BoolVar(&initNeeded, "init", false, "run init procedure for security service.")
	flag.BoolVar(&insecureSkipVerify, "insecureSkipVerify", false, "skip server side SSL verification, mainly for self-signed cert")
	flag.StringVar(&configFileLocation, "configfile", "res/configuration.toml", "configuration file")
	flag.IntVar(&vaultInterval, "vaultInterval", 30, "time to wait between checking Vault status in seconds.")
	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use registry service.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry service.")
	flag.StringVar(&profileDir, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profileDir, "p", "", "Specify a profile other than default.")
	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")

	flag.Usage = usage.HelpCallbackSecuritySecretStore
	flag.Parse()

	params := startup.BootParams{
		UseRegistry: useRegistry,
		ConfigDir:   configDir,
		ProfileDir:  profileDir,
		BootTimeout: internal.BootTimeoutDefault,
	}
	startup.Bootstrap(params, secretstore.Retry, logBeforeInit)
	if secretstore.Configuration == nil {
		// secretstore.LoggingClient wasn't initialized either
		os.Exit(1)
	}

	//step 1: boot up secretstore general steps same as other EdgeX microservice

	//step 2: initialize the communications
	req := secretstore.NewRequester(insecureSkipVerify)
	vaultScheme := secretstore.Configuration.SecretService.Scheme
	vaultHost := fmt.Sprintf("%s:%v", secretstore.Configuration.SecretService.Server, secretstore.Configuration.SecretService.Port)
	intervalDuration := time.Duration(vaultInterval) * time.Second
	vc := secretstoreclient.NewSecretStoreClient(secretstore.LoggingClient, req, vaultScheme, vaultHost)

	//step 3: initialize and unseal Vault
	path := secretstore.Configuration.SecretService.TokenFolderPath
	filename := secretstore.Configuration.SecretService.TokenFile
	absPath := filepath.Join(path, filename)
	for shouldContinue := true; shouldContinue; {
		// Anonymous function used to prevent file handles from accumulating
		func() {
			tokenFile, err := os.Open(absPath)
			if err != nil {
				secretstore.LoggingClient.Error(fmt.Sprintf("unable to open token file at %s%s", path, filename))
			}
			defer tokenFile.Close()
			sCode, _ := vc.HealthCheck()

			switch sCode {
			case http.StatusOK:
				secretstore.LoggingClient.Info(fmt.Sprintf("vault is initialized and unsealed (status code: %d)", sCode))
				shouldContinue = false
			case http.StatusTooManyRequests:
				secretstore.LoggingClient.Error(fmt.Sprintf("vault is unsealed and in standby mode (Status Code: %d)", sCode))
				shouldContinue = false
			case http.StatusNotImplemented:
				secretstore.LoggingClient.Info(fmt.Sprintf("vault is not initialized (status code: %d). Starting initialisation and unseal phases", sCode))
				_, err := vc.Init(secretstore.Configuration.SecretService, tokenFile)
				if err == nil {
					_, err = vc.Unseal(secretstore.Configuration.SecretService, tokenFile)
					if err == nil {
						shouldContinue = false
					}
				}
			case http.StatusServiceUnavailable:
				secretstore.LoggingClient.Info(fmt.Sprintf("vault is sealed (status code: %d). Starting unseal phase", sCode))
				_, err := vc.Unseal(secretstore.Configuration.SecretService, tokenFile)
				if err == nil {
					shouldContinue = false
				}
			default:
				if sCode == 0 {
					secretstore.LoggingClient.Error(fmt.Sprintf("vault is in an unknown state. No Status code available"))
				} else {
					secretstore.LoggingClient.Error(fmt.Sprintf("vault is in an unknown state. Status code: %d", sCode))
				}
			}
		}()

		if shouldContinue {
			secretstore.LoggingClient.Info(fmt.Sprintf("trying Vault init/unseal again in %d seconds", vaultInterval))
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

	//Step 4:
	//TODO: create vault access token for different roles

	//step 5 :
	//TODO: implement credential creation

	cert := secretstore.NewCerts(req, secretstore.Configuration.SecretService.CertPath, absPath)
	existing, err := cert.AlreadyinStore()
	if err != nil {
		secretstore.LoggingClient.Error(err.Error())
		os.Exit(1)
	}

	if existing == true {
		secretstore.LoggingClient.Info("proxy certificate pair are in the secret store already, skip uploading")
		os.Exit(0)
	}

	secretstore.LoggingClient.Info("proxy certificate pair are not in the secret store yet, uploading them")
	cp, err := cert.ReadFrom(secretstore.Configuration.SecretService.CertFilePath, secretstore.Configuration.SecretService.KeyFilePath)
	if err != nil {
		secretstore.LoggingClient.Error("failed to get certificate pair from volume")
		os.Exit(1)
	}

	secretstore.LoggingClient.Info("proxy certificate pair are loaded from volume successfully, will upload to secret store")

	err = cert.UploadToStore(cp)
	if err != nil {
		secretstore.LoggingClient.Error("failed to upload the proxy cert pair into the secret store")
		secretstore.LoggingClient.Error(err.Error())
		os.Exit(1)
	}

	secretstore.LoggingClient.Info("proxy certificate pair are uploaded to secret store successfully, Vault init done successfully")
	os.Exit(0)
}

func logBeforeInit(err error) {
	l := logger.NewClient("security-secretstore-setup", false, "", models.InfoLog)
	l.Error(err.Error())
}
