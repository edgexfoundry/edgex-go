/*
   Copyright 2019 Dell Technologies, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"flag"
	"log"
	"os"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
)

func main() {
	//start := time.Now()
	var configFile string

	flag.StringVar(&configFile, "config", "", "specify JSON configuration file: /path/to/file.json")
	flag.Usage = usage.HelpCallbackSecuritySetup
	flag.Parse()

	if configFile == "" {
		log.Println("ERROR: missing mandatory parameter: -c | --config")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Read the Json config file and unmarshall content into struct type X509Config
	log.Printf("Config file      : %s \n", configFile)
	x509config, err := config.NewX509Config(configFile)
	if err != nil {
		fatalIfErr(err, "Opening configuration file")
	}

	// Create and initialize the fs environment and global vars for the PKI materials
	err = createEnv(x509config)
	if err != nil {
		fatalIfErr(err, "Environment initialization")
	}
}

func createEnv(x509config config.X509Config) error {
	return nil
}

// TODO: ELIMINATE THIS ----------------------------------------------------------
func fatalIfErr(err error, msg string) {
	if err != nil {
		log.Fatalf("ERROR: %s: %s", msg, err) // fatalf() =  Prinf() followed by a call to os.Exit(1)
	}
}
