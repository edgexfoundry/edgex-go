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
	"os"

	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
)

func main() {
	if len(os.Args) < 2 {
		usage.HelpCallbackSecuritySecretStore()
	}

	var initNeeded bool
	var insecureSkipVerify bool
	var configFileLocation string
	var waitInterval int
	var useProfile string
	var useRegistry bool

	flag.BoolVar(&initNeeded, "init", false, "run init procedure for security service.")
	flag.BoolVar(&insecureSkipVerify, "insecureSkipVerify", true, "skip server side SSL verification, mainly for self-signed cert")
	flag.StringVar(&configFileLocation, "configfile", "res/configuration.toml", "configuration file")
	flag.IntVar(&waitInterval, "wait", 30, "time to wait between checking Vault status in seconds.")
	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use registry service.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry service.")
	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")

	flag.Usage = usage.HelpCallbackSecuritySecretStore
	flag.Parse()

	//step 1: boot up secretstore general steps same as other EdgeX microservice

	//step 2: initialize the communications

	//step 3: initialize and unseal Vault

	//Step 4:
	//TODO: create vault access token for different roles

	//step 5 :
	//TODO: implment credential creation

	//step 6:  Push cert key pair for KONG into Vault

}
