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
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/secrets/seed"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

func TestGeneratePrivateKey(t *testing.T) {
	baseSeed := seed.CertificateSeed{
		RSAScheme:  false,
		ECScheme:   false,
		RSAKeySize: seed.RSA_4096,
		ECCurve:    seed.EC_224,
	}

	rsaSeed := baseSeed
	rsaSeed.RSAScheme = true

	ecSeed := baseSeed
	ecSeed.ECScheme = true

	ec256Seed := ecSeed
	ec256Seed.ECCurve = seed.EC_256

	ec384Seed := ecSeed
	ec384Seed.ECCurve = seed.EC_384

	ec521Seed := ecSeed
	ec521Seed.ECCurve = seed.EC_521

	mockLogger := logger.MockLogger{}

	tests := []struct {
		name        string
		seed        seed.CertificateSeed
		expectError bool
	}{
		{"BaseFail", baseSeed, true},
		{"RSAPass", rsaSeed, false},
		{"ECPass", ecSeed, false},
		{"EC256Pass", ec256Seed, false},
		{"EC384Pass", ec384Seed, false},
		{"EC521Pass", ec521Seed, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			private, err := generatePrivateKey(tt.seed, mockLogger)
			if err != nil && !tt.expectError {
				t.Error(err)
				return
			}
			if err == nil {
				if tt.expectError {
					t.Error("expected error but none was thrown")
					return
				}
				// Extract PK from RSA or EC generated SK
				public := private.(crypto.Signer).Public()
				dumpKeyPair(private, mockLogger)
				dumpKeyPair(public, mockLogger)
			}
		})
	}
}

// Note that this only tests for failure because the TestGeneratePrivateKey function above is used to test all the
// success paths
func TestDumpKeyPairFailure(t *testing.T) {
	something := NewFileWriter()
	logger := newMockCryptoLogger(true, t)
	dumpKeyPair(something, logger)
}

// The CryptoLogger mock is used to register unexpected errors with the testing framework.
// See the implementation of the Error() method.
type mockCryptoLogger struct {
	expectError bool
	t           *testing.T
}

func newMockCryptoLogger(expectError bool, t *testing.T) logger.LoggingClient {
	return mockCryptoLogger{expectError: expectError, t: t}
}

// SetLogLevel simulates setting a log severity level
func (lc mockCryptoLogger) SetLogLevel(loglevel string) error {
	return nil
}

// Info simulates logging an entry at the INFO severity level
func (lc mockCryptoLogger) Info(msg string, args ...interface{}) {
}

// Debug simulates logging an entry at the DEBUG severity level
func (lc mockCryptoLogger) Debug(msg string, args ...interface{}) {
}

// Error simulates logging an entry at the ERROR severity level
func (lc mockCryptoLogger) Error(msg string, args ...interface{}) {
	if !lc.expectError {
		lc.t.Error(msg)
		lc.t.Fail()
	}
}

// Trace simulates logging an entry at the TRACE severity level
func (lc mockCryptoLogger) Trace(msg string, args ...interface{}) {
}

// Warn simulates logging an entry at the WARN severity level
func (lc mockCryptoLogger) Warn(msg string, args ...interface{}) {
}
