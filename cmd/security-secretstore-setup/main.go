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
 *
 *******************************************************************************/
package main

import (
	"flag"
	//"fmt"
	//"net/http"
	"os"
	//"path/filepath"
	//"time"

	//"github.com/edgexfoundry/edgex-go/internal"
	//"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	//"github.com/edgexfoundry/edgex-go/internal/security/secretstore"
	//"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	//"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func main() {
	if len(os.Args) < 2 {
		usage.HelpCallbackSecuritySecretStore()
	}

	var initNeeded bool
	var ensureSkipVerify bool
	var configFileLocation string
	var waitInterval int
	var useProfile string
	var useRegistry bool

	flag.BoolVar(&initNeeded, "init", false, "run init procedure for security service.")
	flag.BoolVar(&ensureSkipVerify, "insureskipverify", true, "skip server side SSL verification, mainly for self-signed cert")
	flag.StringVar(&configFileLocation, "configfile", "res/configuration.toml", "configuration file")
	flag.IntVar(&waitInterval, "wait", 30, "time to wait between checking Vault status in seconds.")
	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use registry service.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry service.")
	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")

	flag.Usage = usage.HelpCallbackSecuritySecretStore
	flag.Parse()

	//step 1: boot up secretstore general steps same as other EdgeX microservice
	//params := startup.BootParams{UseRegistry: useRegistry, UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	//startup.Bootstrap(params, secretstore.Retry, logBeforeInit)

	//if initNeeded == false {
	//	secretstore.LoggingClient.Info("skipping initlization and exit. Hint: are you trying to initialize the secret store ? please use the option with --init=true")
	//	os.Exit(0)
	//}

	//step 2: initialize the communications

	/*req := secretstore.NewRequestor(ensureSkipVerify)
	vaultScheme := secretstore.Configuration.SecretService.Scheme
	vaultHost := fmt.Sprintf("%s:%v", secretstore.Configuration.SecretService.Server, secretstore.Configuration.SecretService.Port)
	vc := secretstore.NewVaultClient(req, vaultScheme, vaultHost)
	intervalDuration := time.Duration(waitInterval) * time.Second
	*/

	//step 3: initialize and unseal Vault

	/*loopExit := false

	 for {
		 sCode, _ := vc.HealthCheck()

		 switch sCode {
		 case http.StatusOK:
			 secretstore.LoggingClient.Info(fmt.Sprintf("vault is initialized and unsealed (status code: %d)", sCode))
			 loopExit = true
		 case http.StatusTooManyRequests:
			 secretstore.LoggingClient.Error(fmt.Sprintf("vault is unsealed and in standby mode (Status Code: %d)", sCode))
			 loopExit = true
		 case http.StatusNotImplemented:
			 secretstore.LoggingClient.Info(fmt.Sprintf("vault is not initialized (status code: %d). Starting initialisation and unseal phases", sCode))
			 _, err := vc.Init()
			 if err == nil {
				 _, err = vc.Unseal()
				 if err == nil {
					 loopExit = true
				 }
			 }
		 case http.StatusServiceUnavailable:
			 secretstore.LoggingClient.Info(fmt.Sprintf("vault is sealed (status code: %d). Starting unseal phase", sCode))
			 _, err := vc.Unseal()
			 if err == nil {
				 loopExit = true
			 }
		 default:
			 if sCode == 0 {
				 secretstore.LoggingClient.Error(fmt.Sprintf("vault is in an unknown state. No Status code available"))
			 } else {
				 secretstore.LoggingClient.Error(fmt.Sprintf("vault is in an unknown state. Status code: %d", sCode))
			 }
		 }

		 if loopExit {
			 break
		 }
		 secretstore.LoggingClient.Info(fmt.Sprintf("trying Vault init/unseal again in %d seconds", waitInterval))
		 time.Sleep(intervalDuration)
		}


	 //After vault is init'd and unsealed, it takes a while to get ready to accept any request. During which period any request will get http 500 error.
	 //We need to check the status constantly until it return http StatusOK. The error below shows what may happen if it is not ready.
	 //edgex-vault-worker | INFO: 2018/10/20 10:52:55 Vault has been initialized successfully.
	 //edgex-vault-worker | INFO: 2018/10/20 10:52:55 Vault has been unsealed successfully.
	 //edgex-vault-worker | INFO: 2018/10/20 10:52:55 Reading Admin policy file.
	 //edgex-vault-worker | INFO: 2018/10/20 10:52:55 Importing Vault Admin policy.
	 //edgex-vault-worker | 2018/10/20 10:52:55 ERROR: Import policy failure: admin
	 //edgex-vault-worker | ERROR: 2018/10/20 10:52:55 Import Policy HTTP Status: 500 Internal Server Error (StatusCode: 500)
	 //edgex-vault-worker | ERROR: 2018/10/20 10:52:55 Fatal Error importing Admin policy in Vault.

	 //for {
	//	 if sCode, _ := vc.HealthCheck(); sCode == http.StatusOK {
	//		 break
	//	 }
	 //}
	*/

	//Step 4:
	//TODO: create vault access token for different roles

	//step 5 :
	//TODO: implment credential creation

	//step 6:  Push cert key pair for KONG into Vault

	/* absTokenPath := filepath.Join(secretstore.Configuration.SecretService.TokenFolderPath, secretstore.Configuration.SecretService.TokenFile)
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

	 _, err = cert.UploadToStore(cp)
	 if err != nil {
		 secretstore.LoggingClient.Error("failed to upload the proxy cert pair into the secret store")
		 secretstore.LoggingClient.Error(err.Error())
		 os.Exit(1)
	 }

	 secretstore.LoggingClient.Info("proxy certificate pair are uploaded to secret store successfully, Vault init done successfully")
	 os.Exit(0)
	*/
}

/* func logBeforeInit(err error) {
	 l := logger.NewClient("security-secretstore-setup", false, "", models.InfoLog)
	 l.Error(err.Error())
 }
*/
