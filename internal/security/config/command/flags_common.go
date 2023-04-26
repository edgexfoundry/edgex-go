//
// Copyright (c) 2023 Intel Corporation
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

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/flags"
)

// commonFlags is a custom implementation of flags.Common from go-mod-bootstrap
type commonFlags struct {
	configDir string
}

// NewCommonFlags creates new CommonFlags and initializes it
func NewCommonFlags() flags.Common {
	commonFlags := commonFlags{}
	return &commonFlags
}

// Parse parses the command-line arguments
func (f *commonFlags) Parse(_ []string) {
	flag.StringVar(&f.configDir, "configDir", "", "")
	flag.Usage = HelpCallback

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Please specify command for " + os.Args[0])
		flag.Usage()
		os.Exit(0)
	}

	// Make sure Configuration Provider environment variable isn't set since this service doesn't support using it.
	_ = os.Setenv(internal.ConfigProviderEnvVar, "")
}

// ConfigFileName returns the name of the local configuration file
func (f *commonFlags) ConfigFileName() string {
	return flags.DefaultConfigFile
}

// CommonConfig returns an empty string since common config is not used
func (f *commonFlags) CommonConfig() string {
	return ""
}

// OverwriteConfig returns false since the Configuration provider is not used
func (f *commonFlags) OverwriteConfig() bool {
	return false
}

// UseRegistry returns false since registry is not used
func (f *commonFlags) UseRegistry() bool {
	return false
}

// InDevMode returns false since dev mode is not used
func (f *commonFlags) InDevMode() bool {
	return false
}

// ConfigProviderUrl returns the empty url since Configuration Provider is not used.
func (f *commonFlags) ConfigProviderUrl() string {
	return ""
}

// Profile returns the empty name since profile is not used
func (f *commonFlags) Profile() string {
	return ""
}

// ConfigDirectory returns the directory where the config file(s) are located, if it was specified.
func (f *commonFlags) ConfigDirectory() string {
	return f.configDir
}

// Help displays the usage help message and exit.
func (f *commonFlags) Help() {
	HelpCallback()
}

// HelpCallback displays the help usage message and exits
func HelpCallback() {
	fmt.Printf(
		"Usage: %s [options] <command> [arg...]\n"+
			"Options:\n"+
			"    -h, --help    Show this message\n"+
			"    --configDir     Specify local configuration directory\n"+
			"\n"+
			"Commands:\n"+
			"    help          Show available commands (this text)\n"+
			"    proxy         Configure security settings for EdgeX proxy\n",
		os.Args[0])
}
