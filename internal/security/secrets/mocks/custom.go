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

package mocks

import "github.com/edgexfoundry/edgex-go/internal/security/secrets/config"

func CreateValidX509ConfigMock() config.X509 {
	key := config.KeyScheme{
		DumpKeys:   "false",
		RSA:        "false",
		RSAKeySize: "4096",
		EC:         "true",
		ECCurve:    "384",
	}

	root := config.RootCA{
		CAName:     "EdgeXTrustCA",
		CACountry:  "US",
		CAState:    "CA",
		CALocality: "Austin",
		CAOrg:      "EdgeXFoundry",
	}

	tls := config.TLSServer{
		TLSHost:     "edgex-kong",
		TLSDomain:   "local",
		TLSCountry:  "US",
		TLSSate:     "CA",
		TLSLocality: "San Francisco",
		TLSOrg:      "Kong",
	}

	cfg := config.X509{
		WorkingDir:      ".",
		CreateNewRootCA: "false",
		PKISetupDir:     "testdata",
		DumpConfig:      "false",
		KeyScheme:       key,
		RootCA:          root,
		TLSServer:       tls,
	}
	return cfg
}
