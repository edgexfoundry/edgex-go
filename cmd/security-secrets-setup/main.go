/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2019 Intel Corp.
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

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/security/setup"
	"github.com/edgexfoundry/edgex-go/internal/security/setup/option"
)

type exiter interface {
	exit(int)
}

type exitCode struct{}

type optionDispatcher interface {
	run() (int, error)
}

type pkiInitOptionDispatcher struct{}

var exitInstance = newExit()
var dispatcherInstance = newOptionDispatcher()
var configFile string
var generateOpt bool

func init() {
	// define and register command line flags:
	flag.StringVar(&configFile, "config", "", "specify JSON configuration file: /path/to/file.json")
	flag.StringVar(&configFile, "c", "", "specify JSON configuration file: /path/to/file.json")
	flag.BoolVar(&generateOpt, "generate", false, "to generate a new PKI from scratch")
	flag.Usage = usage.HelpCallbackSecuritySetup
}

func main() {
	start := time.Now()

	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("Please specify option for " + option.SecuritySecretsSetup)
		flag.PrintDefaults()
		exitInstance.exit(0)
		return
	}

	if err := setup.Init(); err != nil {
		// the error returned from Init has already been logged inside the call
		// so here we ignore the error logging
		exitInstance.exit(1)
		return
	}

	if configFile == "" {
		// run with other options for pki-init
		statusCode, err := dispatcherInstance.run()
		if err != nil {
			setup.LoggingClient.Error(err.Error())
		}

		exitInstance.exit(statusCode)
		return
	}

	if err := option.GenTLSAssets(configFile); err != nil {
		setup.LoggingClient.Error(err.Error())
		exitInstance.exit(2)
		return
	}

	setup.LoggingClient.Info(option.SecuritySecretsSetup+" complete", internal.LogDurationKey, time.Since(start).String())
}

func newExit() exiter {
	return &exitCode{}
}

func (code *exitCode) exit(statusCode int) {
	os.Exit(statusCode)
}

func newOptionDispatcher() optionDispatcher {
	return &pkiInitOptionDispatcher{}
}

func (dispatcher *pkiInitOptionDispatcher) run() (statusCode int, err error) {
	opts := option.PkiInitOption{
		GenerateOpt: generateOpt,
	}
	optsExecutor, statusCode, err := option.NewPkiInitOption(opts)
	if err != nil {
		return statusCode, err
	}

	return optsExecutor.ProcessOptions()
}
