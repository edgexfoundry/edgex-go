/*******************************************************************************
 * Copyright 2020 Intel Corp.
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

package flags

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	DefaultConfigProvider = "keeper.http://localhost:59890"
	DefaultConfigFile     = "configuration.yaml"
)

// Common is an interface that defines AP for the common command-line flags used by most EdgeX services
type Common interface {
	OverwriteConfig() bool
	UseRegistry() bool
	InDevMode() bool
	ConfigProviderUrl() string
	Profile() string
	ConfigDirectory() string
	ConfigFileName() string
	CommonConfig() string
	Parse([]string)
	RemoteServiceHosts() []string
	Help()
}

// Default is the Default implementation of Common used by most EdgeX services
type Default struct {
	FlagSet            *flag.FlagSet
	additionalUsage    string
	overwriteConfig    bool
	useRegistry        bool
	devMode            bool
	configProviderUrl  string
	commonConfig       string
	profile            string
	configDir          string
	configFileName     string
	remoteServiceHosts string
}

// NewWithUsage returns a Default struct.
func NewWithUsage(additionalUsage string) *Default {
	return &Default{
		FlagSet:         flag.NewFlagSet("", flag.ExitOnError),
		additionalUsage: additionalUsage,
	}
}

// New returns a Default struct with an empty additional usage string.
func New() *Default {
	return NewWithUsage("")
}

// Parse parses the passed in command-lie arguments looking to the default set of common flags
func (d *Default) Parse(arguments []string) {
	// The flags package doesn't allow for String flags to be specified without a value, so to support
	// -cp/-configProvider without value to indicate using default host value we must detect use of this option with
	// out value and insert the default value before parsing the command line options.
	configProviderRE, _ := regexp.Compile("^--?(cp|configProvider)=?")

	for index, option := range arguments {
		if loc := configProviderRE.FindStringIndex(option); loc != nil {
			if option[loc[1]-1] != '=' {
				arguments[index] = "-cp=" + DefaultConfigProvider
			}

			continue
		}
	}

	// Usage is provided by caller, so leaving individual usage blank here so not confusing where it comes from.
	d.FlagSet.StringVar(&d.configProviderUrl, "configProvider", "", "")
	d.FlagSet.StringVar(&d.configProviderUrl, "cp", "", "")
	d.FlagSet.StringVar(&d.commonConfig, "commonConfig", "", "")
	d.FlagSet.StringVar(&d.commonConfig, "cc", "", "")
	d.FlagSet.BoolVar(&d.overwriteConfig, "overwrite", false, "")
	d.FlagSet.BoolVar(&d.overwriteConfig, "o", false, "")
	d.FlagSet.StringVar(&d.configFileName, "cf", DefaultConfigFile, "")
	d.FlagSet.StringVar(&d.configFileName, "configFile", DefaultConfigFile, "")
	d.FlagSet.StringVar(&d.profile, "profile", "", "")
	d.FlagSet.StringVar(&d.profile, "p", "", ".")
	d.FlagSet.StringVar(&d.configDir, "configDir", "", "")
	d.FlagSet.StringVar(&d.configDir, "cd", "", "")
	d.FlagSet.StringVar(&d.remoteServiceHosts, "remoteServiceHosts", "", "")
	d.FlagSet.StringVar(&d.remoteServiceHosts, "rsh", "", "")
	d.FlagSet.BoolVar(&d.useRegistry, "registry", false, "")
	d.FlagSet.BoolVar(&d.useRegistry, "r", false, "")
	d.FlagSet.BoolVar(&d.devMode, "dev", false, "")
	d.FlagSet.BoolVar(&d.devMode, "d", false, "")

	d.FlagSet.Usage = d.helpCallback

	err := d.FlagSet.Parse(arguments)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

// OverwriteConfig returns whether the local configuration should be pushed (overwrite) into the Configuration provider
func (d *Default) OverwriteConfig() bool {
	return d.overwriteConfig
}

// UseRegistry returns whether the Registry should be used or not
func (d *Default) UseRegistry() bool {
	return d.useRegistry
}

// InDevMode returns whether running in dev mode or not
func (d *Default) InDevMode() bool {
	return d.devMode
}

// ConfigProviderUrl returns the url for the Configuration Provider, if one was specified.
func (d *Default) ConfigProviderUrl() string {
	return d.configProviderUrl
}

// Profile returns the profile name to use, if one was specified
func (d *Default) Profile() string {
	return d.profile
}

// ConfigDirectory returns the directory where the config file(s) are located, if it was specified.
func (d *Default) ConfigDirectory() string {
	return d.configDir
}

// ConfigFileName returns the name of the local configuration file
func (d *Default) ConfigFileName() string {
	return d.configFileName
}

// CommonConfig returns the location for the common configuration
func (d *Default) CommonConfig() string {
	return d.commonConfig
}

func (d *Default) RemoteServiceHosts() []string {
	if len(d.remoteServiceHosts) == 0 {
		return nil
	}

	return strings.Split(d.remoteServiceHosts, ",")
}

// Help displays the usage help message and exit.
func (d *Default) Help() {
	d.helpCallback()
}

// commonHelpCallback displays the help usage message and exits
func (d *Default) helpCallback() {
	fmt.Printf(
		"Usage: %s [options]\n"+
			"Server Options:\n"+
			"    -cp, --configProvider        Indicates to use Configuration Provider service at specified URL.\n"+
			"                                 URL Format: {type}.{protocol}://{host}:{port} ex: keeper.http://localhost:59890\n"+
			"    -cc, --commonConfig          Takes the location where the common configuration is loaded from when\n"+
			"                                 not using the Configuration Provider\n"+
			"    -o, --overwrite              Overwrite configuration in provider with local configuration\n"+
			"                                 *** Use with cation *** Use will clobber existing settings in provider,\n"+
			"                                 problematic if those settings were edited by hand intentionally\n"+
			"    -cf, --configFile <name>     Indicates name of the local configuration file. Defaults to configuration.yaml\n"+
			"    -p, --profile <name>         Indicate configuration profile other than default\n"+
			"    -cd, --configDir             Specify local configuration directory\n"+
			"    -r, --registry               Indicates service should use Registry.\n"+
			"    -rsh, \n"+
			"     --remoteServiceHosts \n"+
			"          <host names>           Indicates that the service is running remote from the core EdgeX services and\n"+
			"                                 to use the listed host names to connect remotely. <host names> contains 3 comma seperated host names seperated by ','.\n"+
			"                                 1st is the local system host name, 2nd is the remote system host name and 3rd is the WebServer bind host name\n"+
			"                                 example: -rsh=192.0.1.20,192.0.1.5,localhost\n"+
			"    -d, --dev                    Indicates service to run in developer mode which causes Host configuration values to be overridden.\n"+
			"                                 with `localhost`. This is so that it will run with other services running in Docker (aka hybrid mode)\n"+
			"%s\n"+
			"Common Options:\n"+
			"    -h, --help                   Show this message\n",
		os.Args[0], d.additionalUsage,
	)
	os.Exit(0)
}
