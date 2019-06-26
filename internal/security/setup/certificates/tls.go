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

package certificates

import (
	"crypto"
	"crypto/x509"

	"github.com/edgexfoundry/edgex-go/internal/security/setup"
)

type tlsGenerator struct {
}

func newTlsGenerator(seed setup.CertificateSeed) (gen tlsGenerator, err error) {
	return tlsGenerator{}, nil
}

func (gen tlsGenerator) Generate() (*x509.Certificate, crypto.PrivateKey, error) {
	return &x509.Certificate{}, nil, nil
}
