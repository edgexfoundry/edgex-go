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

package identity

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
)

// AssembleServerChains takes in an array of certificates, finds all certificates with
// x509.ExtKeyUsageAny or x509.ExtKeyUsageServerAuth and builds an array of leaf-first
// chains. Chains are built starting from server authentication certificates found in `certs`
// and the signer chains are built from `certs` and `cas`. Both slices are de-duped
// and the `cas` slice is filtered for certificates with the CA flag set.
func AssembleServerChains(certs []*x509.Certificate, cas []*x509.Certificate) ([][]*x509.Certificate, error) {
	if len(certs) == 0 {
		return nil, nil
	}

	certs = getUniqueCerts(certs)

	var chains [][]*x509.Certificate

	var serverCerts []*x509.Certificate

	for _, cert := range certs {
		//if we find CAs add them to the known CAs
		if cert.IsCA {
			cas = append(cas, cert)
		}
		//if not a CA and have any DNS/IP SANs
		//primary lookup method is "a leaf with SANs"
		if (!cert.IsCA) && (len(cert.DNSNames) != 0 || len(cert.IPAddresses) != 0) {
			serverCerts = append(serverCerts, cert)
		} else {
			//check for key usage, a CA may be a server certificate
			//this is a secondary check as extended key usages aren't always respected or used
			for _, usage := range cert.ExtKeyUsage {
				if usage == x509.ExtKeyUsageAny || usage == x509.ExtKeyUsageServerAuth {
					serverCerts = append(serverCerts, cert)
				}
			}
		}
	}

	cas = getUniqueCas(cas)

	for _, serverCert := range serverCerts {
		chain := buildChain(serverCert, cas)
		chains = append(chains, chain)
	}

	return chains, nil
}

// buildChain will build as much of a chain as possible from startingLeaf up using signature checking.
func buildChain(startingLeaf *x509.Certificate, cas []*x509.Certificate) []*x509.Certificate {
	var chain []*x509.Certificate

	current := startingLeaf

	for current != nil {
		chain = append(chain, current)

		//check to see if we are the root
		if current.IsCA {
			if err := current.CheckSignatureFrom(current); err == nil {
				break
			}
		}

		parentFound := false

		//search by checking signature
		for _, next := range cas {
			if next.IsCA {
				if err := current.CheckSignatureFrom(next); err == nil {
					current = next
					parentFound = true
					break
				}
			}
		}

		if !parentFound {
			current = nil
		}
	}

	return chain
}

// ChainsToTlsCerts converts and array of x509 certificate chains to an array of tls.Certificates (which
// have their own internal arrays of raw certificates). It is assumed the same private key is used for
// all chains.
func ChainsToTlsCerts(chains [][]*x509.Certificate, key crypto.PrivateKey) []*tls.Certificate {
	tlsCerts := make([]*tls.Certificate, len(chains))

	for chainIdx, chain := range chains {
		tlsCerts[chainIdx] = &tls.Certificate{
			Certificate: make([][]byte, len(chain)),
			Leaf:        chain[0],
			PrivateKey:  key,
		}

		for certIdx, cert := range chain {
			tlsCerts[chainIdx].Certificate[certIdx] = cert.Raw
		}
	}

	return tlsCerts
}
