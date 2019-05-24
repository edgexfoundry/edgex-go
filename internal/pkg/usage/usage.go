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

// usage will print out the flag options for the server.
func HelpCallback() {
	msg := fmt.Sprintf(usageStr, os.Args[0])
	fmt.Printf("%s\n", msg)
	os.Exit(0)
}

func HelpCallbackConfigSeed() {
	msg := fmt.Sprintf(configSeedUsageStr, os.Args[0])
	fmt.Printf("%s\n", msg)
	os.Exit(0)
}
