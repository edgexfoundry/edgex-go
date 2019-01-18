/*
 * Copyright 2019 ObjectBox Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/edgexfoundry/edgex-go/cmd/set-db/templates"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	var service string
	var providerImport string

	flag.StringVar(&service, "service", service, "[required] The name of the service that should be configured.")
	flag.StringVar(&providerImport, "provider", providerImport, "[required] Import path for the database provider implementation.")
	flag.Parse()

	if service == "" {
		showUsageAndExit()
	}

	log.Printf("Configuring %s database", service)

	if err := createInitDb(service, configuration{ProviderImport: providerImport}); err != nil {
		log.Fatal(err)
	}
}

func showUsageAndExit() {
	flag.Usage()
	os.Exit(1)
}

// the following code should be moved to a new internal package

const filePerms = 0644

type configuration struct {
	ProviderImport string
	Service        struct {
		Package       string
		Path          string
		InterfacePath string
		Interface     string
	}
}

// mapping between service-name and internal/service/name
// a naive implementation just replaces the "-", works for the currently available services
func servicePath(service string) string {
	return "internal/" + strings.Replace(service, "-", "/", -1)
}

// creates initdb.go for the service
func createInitDb(service string, conf configuration) error {
	// fill in the missing configuration
	conf.Service.Path = servicePath(service)
	conf.Service.Package = path.Base(conf.Service.Path)

	// export-* has different interface path
	if strings.HasPrefix(service, "export") {
		conf.Service.InterfacePath = path.Dir(conf.Service.Path)
	} else {
		conf.Service.InterfacePath = path.Join(conf.Service.Path, "interfaces")
	}

	conf.Service.Interface = path.Base(conf.Service.InterfacePath)

	// generate the file contents
	var err, err2 error
	var file = filepath.Join(conf.Service.Path, "initdb.go")
	var contents []byte
	if contents, err = generateInitDb(conf); err != nil {
		return fmt.Errorf("can't generate service DB initialization file %s: %s", file, err)
	}

	// format it
	if formattedSource, err := format.Source(contents); err != nil {
		// we just store error but still write the file so that we can check it manually
		err2 = fmt.Errorf("failed to format the service DB initialization file %s: %s", file, err)
	} else {
		contents = formattedSource
	}

	// write the file to the service package
	if err = ioutil.WriteFile(file, contents, filePerms); err != nil {
		return fmt.Errorf("can't write the service DB initialization file %s: %s", file, err)
	} else if err2 != nil {
		// now when the model has been written (for debugging purposes), we can return the error
		return err2
	}

	log.Printf("Service %s is now using %s as a persistence provider", service, conf.ProviderImport)
	return nil
}

// generates the source code for the initdb.go file
func generateInitDb(conf configuration) (data []byte, err error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	if err = templates.InitDb.Execute(writer, conf); err != nil {
		return nil, fmt.Errorf("template execution failed: %s", err)
	}

	if err = writer.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush buffer: %s", err)
	}

	return b.Bytes(), nil
}
