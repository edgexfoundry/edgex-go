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
	"testing"

	secretsConfig "github.com/edgexfoundry/edgex-go/internal/security/secrets/config"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/contract"

	"github.com/edgexfoundry/edgex-go/internal/security/secrets/mocks"
	"github.com/edgexfoundry/edgex-go/internal/security/secrets/seed"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

func TestRootCertGenerate(t *testing.T) {
	writer := mockFileWriter{}
	mockLogger := logger.MockLogger{}
	cfg := mocks.CreateValidX509ConfigMock()
	dir := createDirectoryHandlerMock(cfg, t)
	certificateSeed, err := seed.NewCertificateSeed(cfg, dir)
	if err != nil {
		t.Error(err.Error())
		return
	}

	seedGeneratorOn := certificateSeed
	seedGeneratorOn.NewCA = true
	seedGeneratorOn.DumpKeys = true

	schemesOff := seedGeneratorOn
	schemesOff.ECScheme = false
	schemesOff.RSAScheme = false

	tests := []struct {
		name            string
		certificateSeed seed.CertificateSeed
		expectError     bool
	}{
		{"GenOffOK", certificateSeed, false},
		{"GenOnOK", seedGeneratorOn, false},
		{"SchemeFail", schemesOff, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := NewCertificateGenerator(RootCertificate, tt.certificateSeed, writer, mockLogger)
			if generator != nil {
				err = generator.Generate()
			}
			if err != nil && !tt.expectError {
				t.Error(err)
				return
			}
			if err == nil && tt.expectError {
				t.Error("expected error but none was thrown")
				return
			}
		})
	}
}

func createDirectoryHandlerMock(cfg secretsConfig.X509, t *testing.T) contract.DirectoryHandler {
	dir, err := cfg.PkiCADir()
	if err != nil {
		t.Error(err.Error())
		return nil
	}
	mock := mocks.DirectoryHandler{}
	mock.On("Create", dir).Return(nil)
	mock.On("Verify", dir).Return(nil)
	return &mock
}
