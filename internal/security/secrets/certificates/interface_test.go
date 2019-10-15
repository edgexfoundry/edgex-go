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
	"os"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/secrets"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

func TestNewCertificateGenerator(t *testing.T) {
	seed := secrets.CertificateSeed{}
	writer := mockFileWriter{}
	mockLogger := logger.MockLogger{}

	tests := []struct {
		name        string
		certType    CertificateType
		expectError bool
	}{
		{"RootCAPass", RootCertificate, false},
		{"TLSPass", TLSCertificate, false},
		{"InvalidFail", 3, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCertificateGenerator(tt.certType, seed, writer, mockLogger)
			if err != nil && !tt.expectError {
				t.Errorf("unexpected error %v", err)
				return
			}
			if err == nil && tt.expectError {
				t.Error("expected error, none returned")
				return
			}
		})
	}
}

type mockFileWriter struct{}

func (w mockFileWriter) Write(fileName string, contents []byte, permissions os.FileMode) error {
	return nil
}
