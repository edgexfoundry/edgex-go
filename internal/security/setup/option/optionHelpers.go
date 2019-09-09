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

package option

import (
	"errors"

	config "github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/security/setup"
	"github.com/edgexfoundry/edgex-go/internal/security/setup/certificates"
)

// GenTLSAssets generates the TLS assets based on the JSON configuration file
func GenTLSAssets(jsonConfig string) error {
	// Read the Json config file and unmarshall content into struct type X509Config
	x509Config, err := config.NewX509Config(jsonConfig)
	if err != nil {
		return err
	}

	if setup.NewDirectoryHandler(setup.LoggingClient) == nil {
		return errors.New("setup.LoggingClient is nil")
	}

	seed, err := setup.NewCertificateSeed(x509Config, setup.NewDirectoryHandler(setup.LoggingClient))
	if err != nil {
		return err
	}

	rootCA, err := certificates.NewCertificateGenerator(certificates.RootCertificate, seed, certificates.NewFileWriter(), setup.LoggingClient)
	if err != nil {
		return err
	}

	err = rootCA.Generate()
	if err != nil {
		return err
	}

	tlsCert, err := certificates.NewCertificateGenerator(certificates.TLSCertificate, seed, certificates.NewFileWriter(), setup.LoggingClient)
	if err != nil {
		return err
	}

	tlsCert.Generate()
	if err != nil {
		return err
	}

	return nil
}
