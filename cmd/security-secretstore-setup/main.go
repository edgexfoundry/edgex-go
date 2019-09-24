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
	"os"
	"path/filepath"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/security/secretstore"
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
	var waitInterval int
	var configDir, profileDir string
	var useRegistry bool

	flag.BoolVar(&initNeeded, "init", false, "run init procedure for security service.")
	flag.BoolVar(&insecureSkipVerify, "insecureSkipVerify", false, "skip server side SSL verification, mainly for self-signed cert")
	flag.StringVar(&configFileLocation, "configfile", "res/configuration.toml", "configuration file")
	flag.IntVar(&waitInterval, "wait", 30, "time to wait between checking Vault status in seconds.")
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

	//step 1: boot up secretstore general steps same as other EdgeX microservice

	//step 2: initialize the communications
	req := secretstore.NewRequester(insecureSkipVerify)

	//step 3: initialize and unseal Vault

	//Step 4:
	//TODO: create vault access token for different roles

	//step 5 :
	//TODO: implement credential creation

	absTokenPath := filepath.Join(secretstore.Configuration.SecretService.TokenFolderPath, secretstore.Configuration.SecretService.TokenFile)
	cert := secretstore.NewCerts(req, secretstore.Configuration.SecretService.CertPath, absTokenPath)
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
