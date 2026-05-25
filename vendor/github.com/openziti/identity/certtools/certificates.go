/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package certtools

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func LoadCert(pemBytes []byte) ([]*x509.Certificate, error) {
	certs := make([]*x509.Certificate, 0)
	var keyBlock *pem.Block
	for len(pemBytes) > 0 {
		keyBlock, pemBytes = pem.Decode(pemBytes)

		//keyBlock is nil if there are no subsequent PEM blocks
		//non-PEM data before blocks id treated as headers
		//trailing data after last PEM (or when there are no PEMs) is ignored
		if keyBlock == nil {
			break
		}

		switch keyBlock.Type {
		case "CERTIFICATE":
			if c, err := x509.ParseCertificate(keyBlock.Bytes); err == nil {
				certs = append(certs, c)
			}
		}
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificate found")
	}

	return certs, nil
}

func LoadCertFromFile(f string) ([]*x509.Certificate, error) {
	if pemBytes, err := os.ReadFile(f); err != nil {
		return nil, err
	} else {
		return LoadCert(pemBytes)
	}
}
