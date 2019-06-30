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
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/setup"
	"github.com/edgexfoundry/edgex-go/internal/security/setup/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

func TestTLSCertGenerate(t *testing.T) {
	writer := mockFileWriter{}
	logger := logger.MockLogger{}
	cfg := mocks.CreateValidX509ConfigMock()
	dir := createDirectoryHandlerMock(cfg, t)
	seed, err := setup.NewCertificateSeed(cfg, dir)
	if err != nil {
		t.Error(err.Error())
		return
	}

	dumpKeysOn := seed
	dumpKeysOn.DumpKeys = true

	schemesOff := seed
	schemesOff.ECScheme = false
	schemesOff.RSAScheme = false

	certFileNotFound := seed
	certFileNotFound.CACertFile = overridePath("blank.pem", certFileNotFound.CACertFile)

	certFileInvalid := seed
	certFileInvalid.CACertFile = overridePath("EdgeXTrustCAInvalid.pem", certFileInvalid.CACertFile)

	keyfileNotFound := seed
	keyfileNotFound.CAKeyFile = overridePath("blank.priv.key", certFileNotFound.CAKeyFile)

	keyFileInvalid := seed
	keyFileInvalid.CAKeyFile = overridePath("EdgeXTrustCAInvalid.priv.key", keyFileInvalid.CAKeyFile)

	tests := []struct {
		name        string
		seed        setup.CertificateSeed
		expectError bool
	}{
		{"DefaultConfigOK", seed, false},
		{"DefaultWithDumpKeys", dumpKeysOn, false},
		{"SchemeFail", schemesOff, true},
		{"CertFileNotFound", certFileNotFound, true},
		{"CertFileInvalid", certFileInvalid, true},
		{"KeyFileNotFound", keyfileNotFound, true},
		{"KeyFileInvalid", keyFileInvalid, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := NewCertificateGenerator(TLSCertificate, tt.seed, writer, logger)
			if generator != nil {
				err = generator.Generate()
			}
			if err != nil && !tt.expectError {
				t.Error(err)
			}
			if err == nil && tt.expectError {
				t.Error("expected error but none was thrown")
			}
		})
	}
}

func overridePath(fileName string, path string) string {
	idx := strings.LastIndex(path, "/")
	stem := path[:idx+1]
	return stem + fileName
}
