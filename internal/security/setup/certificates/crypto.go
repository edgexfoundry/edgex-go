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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/security/setup"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// generatePrivateKey creates a new RSA or EC based private key (sk)
// ----------------------------------------------------------
func generatePrivateKey(seed setup.CertificateSeed, logger logger.LoggingClient) (crypto.PrivateKey, error) {

	if seed.RSAScheme {
		logger.Debug(fmt.Sprintf("Generating private key with RSA scheme %v", seed.RSAKeySize))
		return rsa.GenerateKey(rand.Reader, int(seed.RSAKeySize))
	}

	if seed.ECScheme {
		logger.Debug(fmt.Sprintf("Generating private key with EC scheme %v", seed.ECCurve))
		switch seed.ECCurve {
		case setup.EC_224: // secp224r1 NIST P-224
			return ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
		case setup.EC_256: // secp256v1 NIST P-256
			return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		case setup.EC_384: // secp384r1 NIST P-384
			return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		case setup.EC_521: // secp521r1 NIST P-521
			return ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		}
	}

	return nil, fmt.Errorf("Unknown key scheme: RSA[%t] EC[%t]", seed.RSAScheme, seed.ECScheme)
}

// dumpKeyPair output sk,pk keypair (RSA or EC) to console
// !!! Debug only for obvious security reasons...
// ----------------------------------------------------------
func dumpKeyPair(key interface{}, logger logger.LoggingClient) {
	switch key.(type) {
	case *rsa.PrivateKey:
		logger.Debug(fmt.Sprintf(">> RSA SK: %q", key))
	case *ecdsa.PrivateKey:
		logger.Debug(fmt.Sprintf(">> ECDSA SK: %q", key))
	case *rsa.PublicKey:
		logger.Debug(fmt.Sprintf(">> RSA PK: %q", key))
	case *ecdsa.PublicKey:
		logger.Debug(fmt.Sprintf(">> ECDSA PK: %q", key))
	default:
		logger.Error("Unsupported Key Type")
	}
}
