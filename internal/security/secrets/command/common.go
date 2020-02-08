//
// Copyright (c) 2020 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/contract"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

// CommonFlags is a custom implementation of commandline.CommonFlags from go-mod-bootstrap
type CommonFlags struct {
	configDir string
}

// NewCommonFlags creates new CommonFlags and initializes it
func NewCommonFlags() *CommonFlags {
	commonFlags := CommonFlags{}
	return &commonFlags
}

// ConfigFileName returns the name of the local configuration file
func (f *CommonFlags) ConfigFileName() string {
	return internal.ConfigFileName
}

// Parse parses the command-line arguments
func (f *CommonFlags) Parse(_ []string) {
	flag.StringVar(&f.configDir, "confdir", "", "")
	flag.Usage = helpCallback

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Please specify subcommand for " + clients.SecuritySecretsSetupServiceKey)
		flag.Usage()
		os.Exit(contract.StatusCodeExitNormal)
	}

	// Make sure Configuration Provider environment variable isn't set since this service doesn't support using it.
	_ = os.Setenv(internal.ConfigProviderEnvVar, "")
}

// UseRegistry returns false since registry is not used
func (f *CommonFlags) UseRegistry() bool {
	return false
}

// ConfigProviderUrl returns the empty url since Configuration Provider is not used.
func (f *CommonFlags) ConfigProviderUrl() string {
	return ""
}

// Profile returns the empty name since profile is not used
func (f *CommonFlags) Profile() string {
	return ""
}

// ConfigDirectory returns the directory where the config file(s) are located, if it was specified.
func (f *CommonFlags) ConfigDirectory() string {
	return f.configDir
}

// Help displays the usage help message and exit.
func (f *CommonFlags) Help() {
	helpCallback()
}

// helpCallback displays the help usage message and exits
func helpCallback() {
	fmt.Printf(
		"Usage: %s <subcommands> [options]\n"+
			"Server Options:\n"+
			"    --confdir                       Specify local configuration directory\n"+
			"\n"+
			"Server Subcommand:\n"+
			"	generate                        Generate PKI afresh every time and deployed. Typically, this will be whenever the framework is started.\n"+
			"	cache                           Generate PKI exactly once and then copied to a designated cache location for future use.  The PKI is then deployed from the cached location.\n"+
			"	import                          Import PKI from cached location to deployed location.  It requires PKI assets to be pre-populated first and it raises an error if the PKI assets in the cached location are empty.\n"+
			"	legacy                          [Will be deprecated] Legacy mode to generate PKI\n"+
			"	  -c, --config <name>           Provide absolute path to JSON configuration file\n"+
			"\n"+
			"Common Options:\n"+
			"    -h, --help                      Show this message",
		os.Args[0])
	os.Exit(0)
}
