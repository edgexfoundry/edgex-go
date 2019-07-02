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
 *******************************************************************************/

package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/security/setup"
	"github.com/edgexfoundry/edgex-go/internal/security/setup/certificates"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

func main() {
	start := time.Now()
	var configFile string

	flag.StringVar(&configFile, "config", "", "specify JSON configuration file: /path/to/file.json")
	flag.StringVar(&configFile, "c", "", "specify JSON configuration file: /path/to/file.json")
	flag.Usage = usage.HelpCallbackSecuritySetup
	flag.Parse()

	if configFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	setup.Init()

	// Create and initialize the fs environment and global vars for the PKI materials
	lc := logger.NewClient("security-secrets-setup", setup.Configuration.Logging.EnableRemote,
		setup.Configuration.Logging.File, setup.Configuration.Writable.LogLevel)

	// Read the Json config file and unmarshall content into struct type X509Config
	x509config, err := config.NewX509Config(configFile)
	if err != nil {
		lc.Error(err.Error())
		return
	}

	seed, err := setup.NewCertificateSeed(x509config, setup.NewDirectoryHandler(lc))
	if err != nil {
		lc.Error(err.Error())
		return
	}

	rootCA, err := certificates.NewCertificateGenerator(certificates.RootCertificate, seed, certificates.NewFileWriter(), lc)
	if err != nil {
		lc.Error(err.Error())
		return
	}

	err = rootCA.Generate()
	if err != nil {
		lc.Error(err.Error())
		return
	}

	tlsCert, err := certificates.NewCertificateGenerator(certificates.TLSCertificate, seed, certificates.NewFileWriter(), lc)
	if err != nil {
		lc.Error(err.Error())
		return
	}

	tlsCert.Generate()
	if err != nil {
		lc.Error(err.Error())
		return
	}
	lc.Info("PKISetup complete", internal.LogDurationKey, time.Since(start).String())
}

// TODO: ELIMINATE THIS ----------------------------------------------------------
func fatalIfErr(err error, msg string) {
	if err != nil {
		log.Fatalf("ERROR: %s: %s", msg, err) // fatalf() =  Prinf() followed by a call to os.Exit(1)
	}
}
