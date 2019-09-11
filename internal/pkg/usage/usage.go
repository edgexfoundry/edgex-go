/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
 *******************************************************************************/
package usage

import (
	"fmt"
	"os"
)

var usageStr = `
Usage: %s [options]
Server Options:
    -r, --registry                  Indicates service should use Registry
    -p, --profile <name>            Indicate configuration profile other than default
    --confdir                       Specify local configuration directory
Common Options:
    -h, --help                      Show this message
`

var configSeedUsageStr = `
Usage: %s [options]
Server Options:
    -c, --cmd <dir>                 Provide absolute path to "cmd" directory containing EdgeX service configuration
    -o, --overwrite                 Indicates service should overwrite any entries already present in the configuration
    -p, --profile <name>            Indicate configuration profile other than default
    -r, --props <dir>               Provide alternate location for legacy application.properties files
    --confdir                       Specify local configuration directory
    
Common Options:
    -h, --help                      Show this message
`

var securitySetupUsageStr = `
Usage: %s [options]
Server Options:
    -c, --config <name>             Provide absolute path to JSON configuration file

Server Subcommand:	
    generate                        Generate PKI afresh every time and deployed. Typically, this will be whenever the framework is started.

Common Options:
    -h, --help                      Show this message
`

var securityProxySetupUsageStr = `
Usage: %s [options]
Server Options:	
    -p, --profile <name>            Indicate configuration profile other than default
    -r, --registry                  Indicates service should use Registry
	--insureskipverify=true/false   Indicates if skipping the server side SSL cert verifcation, similar to -k of curl
	--init=true/false               Indicates if security service should be initialized
	--reset=true/false              Indicate if security service should be reset to initialization status
	--useradd=<username>            Create an account and return JWT
	--group=<groupname>             Group name the user belongs to
	--userdel=<username>            Delete an account		
	--configfile=<file.toml>        Use a different config file (default: res/configuration.toml)

Common Options:
	-h, --help                      Show this message
`

var securitySecretStoreSetupUsageStr = `
Usage: %s [options]
Server Options:
	-p, --profile <name>            Indicate configuration profile other than default
	-r, --registry                  Indicates service should use Registry
	--insecureSkipVerify=true/false			Indicates if skipping the server side SSL cert verifcation, similar to -k of curl
	--init=true/false				Indicates if security service should be initialized
	--configfile=<file.toml>			Use a different config file (default: res/configuration.toml)
	--wait=<time in seconds>		Indicates how long the program will pause between the vault initialization until it succeeds
Common Options:
	-h, --help					Show this message
`

// usage will print out the flag options for the server.
func HelpCallback() {
	fmt.Printf(usageStr, os.Args[0])
	os.Exit(0)
}

func HelpCallbackConfigSeed() {
	fmt.Printf(configSeedUsageStr, os.Args[0])
	os.Exit(0)
}

func HelpCallbackSecuritySetup() {
	fmt.Printf(securitySetupUsageStr, os.Args[0])
}

func HelpCallbackSecurityProxy() {
	fmt.Printf(securityProxySetupUsageStr, os.Args[0])
	os.Exit(0)
}

func HelpCallbackSecuritySecretStore() {
	msg := fmt.Sprintf(securitySecretStoreSetupUsageStr, os.Args[0])
	fmt.Printf("%s\n", msg)
	os.Exit(0)
}
