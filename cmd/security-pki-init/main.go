//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	option "github.com/edgexfoundry/edgex-go/internal/security/pkiinit/option"
)

type exit interface {
	callExit(int)
}

type exitCode struct{}

type optionDispatcher interface {
	run() (int, error)
}

type pkiInitOptionDispatcher struct{}

var exitInstance = newExit()
var dispatcherInstance = newOptionDispatcher()
var helpOpt bool
var generateOpt bool
var importOpt bool
var cacheOpt bool
var cacheCAOpt bool

func init() {
	// define and register command line flags:
	flag.BoolVar(&helpOpt, "h", false, "help message")
	flag.BoolVar(&helpOpt, "help", false, "help message")
	flag.BoolVar(&generateOpt, "generate", false, "to generate a new PKI from scratch")
	flag.BoolVar(&importOpt, "import", false, " deploys a PKI from a cached PKI")
	flag.BoolVar(&cacheOpt, "cache", false, "generates a fresh PKI exactly once then caches it for future use")
	flag.BoolVar(&cacheCAOpt, "cacheca", false, "generates a fresh PKI exactly once then caches it with the CA keys presevered")
}

func main() {
	flag.Parse()

	if helpOpt {
		// as specified in the requirement, help option terminates with 0 exit status
		flag.Usage()
		exitInstance.callExit(0)
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Please specify option for pki-init.")
		flag.Usage()
		exitInstance.callExit(1)
		return
	}

	statusCode, err := dispatcherInstance.run()

	if err != nil {
		log.Println(err)
	}

	exitInstance.callExit(statusCode)
}

func newExit() exit {
	return &exitCode{}
}

func newOptionDispatcher() optionDispatcher {
	return &pkiInitOptionDispatcher{}
}

func (code *exitCode) callExit(statusCode int) {
	os.Exit(statusCode)
}

func setupPkiInitOption() (executor option.OptionsExecutor, status int, err error) {
	opts := option.PkiInitOption{
		GenerateOpt: generateOpt,
		ImportOpt:   importOpt,
		CacheOpt:    cacheOpt,
		CacheCAOpt:  cacheCAOpt,
	}

	return option.NewPkiInitOption(opts)
}

func (dispatcher *pkiInitOptionDispatcher) run() (statusCode int, err error) {

	optsExecutor, statusCode, err := setupPkiInitOption()
	if err != nil {
		return
	}

	return optsExecutor.ProcessOptions()
}
