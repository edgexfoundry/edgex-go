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

package setup

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/security/setup/mocks"
)

func TestNewCertificateSeed(t *testing.T) {
	valid := createValidConfig()

	validNewCA := valid
	validNewCA.CreateNewRootCA = "true"

	validDumpConfig := valid
	validDumpConfig.DumpConfig = "true"

	validNonLocal := valid
	validNonLocal.TLSServer.TLSDomain = "abc.com"

	caInvalid := valid
	caInvalid.CreateNewRootCA = "test"

	dumpKeysInvalid := valid
	dumpKeysInvalid.KeyScheme.DumpKeys = "test"

	rsaInvalid := valid
	rsaInvalid.KeyScheme.RSA = "test"

	keySizeParseInvalid := valid
	keySizeParseInvalid.KeyScheme.RSAKeySize = "test"

	keySizeValueInvalid := valid
	keySizeValueInvalid.KeyScheme.RSAKeySize = "678"

	ecInvalid := valid
	ecInvalid.KeyScheme.EC = "test"

	curveParseInvalid := valid
	curveParseInvalid.KeyScheme.ECCurve = "test"

	curveValueInvalid := valid
	curveValueInvalid.KeyScheme.ECCurve = "555"

	dumpConfigInvalid := valid
	dumpConfigInvalid.DumpConfig = "test"

	tests := []struct {
		name        string
		dir         DirectoryHandler
		cfg         config.X509Config
		expectError bool
	}{
		{"Pass", createDirectoryHandlerMock(valid, t), valid, false},
		{"NewCAPass", createDirectoryHandlerMock(validNewCA, t), validNewCA, false},
		{"DumpConfigPass", createDirectoryHandlerMock(validDumpConfig, t), validDumpConfig, false},
		{"NonLocalPass", createDirectoryHandlerMock(validNonLocal, t), validNonLocal, false},
		{"CAParseFail", createDirectoryHandlerMock(caInvalid, t), caInvalid, true},
		{"DumpKeysFail", createDirectoryHandlerMock(dumpKeysInvalid, t), dumpKeysInvalid, true},
		{"RSAParseFail", createDirectoryHandlerMock(rsaInvalid, t), rsaInvalid, true},
		{"KeySizeParseFail", createDirectoryHandlerMock(keySizeParseInvalid, t), keySizeParseInvalid, true},
		{"KeySizeValueFail", createDirectoryHandlerMock(keySizeValueInvalid, t), keySizeValueInvalid, true},
		{"ECParseFail", createDirectoryHandlerMock(ecInvalid, t), ecInvalid, true},
		{"CurveParseFail", createDirectoryHandlerMock(curveParseInvalid, t), curveParseInvalid, true},
		{"CurveValueFail", createDirectoryHandlerMock(curveValueInvalid, t), curveValueInvalid, true},
		{"DumpConfigFail", createDirectoryHandlerMock(dumpConfigInvalid, t), dumpConfigInvalid, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCertificateSeed(tt.cfg, tt.dir)
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

func createDirectoryHandlerMock(cfg config.X509Config, t *testing.T) DirectoryHandler {
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

func createValidConfig() config.X509Config {
	key := config.KeyScheme{
		DumpKeys:   "false",
		RSA:        "false",
		RSAKeySize: "4096",
		EC:         "true",
		ECCurve:    "384",
	}

	root := config.RootCA{
		CAName:     "EdgeXFoundryCA",
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

	cfg := config.X509Config{
		WorkingDir:      ".",
		CreateNewRootCA: "false",
		PKISetupDir:     "pki",
		DumpConfig:      "false",
		KeyScheme:       key,
		RootCA:          root,
		TLSServer:       tls,
	}
	return cfg
}
