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
 * @author: Tingyu Zeng, Dell
 * @version: 1.1.0
 *******************************************************************************/
package proxy

import (
	"fmt"
	"os"
)

var usageStr = `
Usage: %s [options]
Server Options:	
	--insureskipverify=true/false			Indicates if skipping the server side SSL cert verifcation, similar to -k of curl
	--init=true/false				Indicates if security service should be initialized
	--reset=true/false				Indicate if security service should be reset to initialization status
	--useradd=<username>				Create an account and return JWT
	--group=<groupname>					Group name the user belongs to
	--userdel=<username>				Delete an account		
	--configfile=<file.toml>			Use a different config file (default: res/configuration.toml)
	Common Options:
	-h, --help					Show this message
`

// 	Print out the flag options for the server.
func HelpCallback() {
	msg := fmt.Sprintf(usageStr, os.Args[0])
	fmt.Printf("%s\n", msg)
	os.Exit(0)
}
