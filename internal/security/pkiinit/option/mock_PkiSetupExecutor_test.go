//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//
package option

import (
	"fmt"
	"os"
	"path/filepath"

	cert "github.com/edgexfoundry/edgex-go/internal/security/pkiinit/cert"
)

// Note: this is a test helper

type mockPkiSetupRunner struct {
	pkiSetupRunner
}

func (mockPkiSetup *mockPkiSetupRunner) checkRunConfig() error {
	return nil
}

func (mockPkiSetup *mockPkiSetupRunner) call() error {
	jsonPath := mockPkiSetup.pkiSetupRunner.runConfig.pkiSetupVaultJSONPath
	reader := cert.NewX509ConfigReader(jsonPath)
	config, readErr := reader.Read()
	if readErr != nil {
		return readErr
	}
	pkiOutputDir, err := config.GetPkiOutputDir()
	fmt.Println("In test, pkiOutDir: ", pkiOutputDir)

	if err != nil {
		return err
	}
	if err := createDirectoryIfNotExists(pkiOutputDir); err != nil {
		return err
	}

	// then copy ca and vault tls testing materials onto the folder
	curDir, _ := os.Getwd()
	testDataDir := filepath.Join(curDir, "..", "testdata")
	fmt.Println("testDataDir ", testDataDir)
	testCaCertFile := filepath.Join(testDataDir, config.GetCAPemFileName())
	testCaKeyFile := filepath.Join(testDataDir, config.GetCAPrivateKeyFileName())
	testVaultCertFile := filepath.Join(testDataDir, config.GetTLSPemFileName())
	testVaultKeyFile := filepath.Join(testDataDir, config.GetTLSPrivateKeyFileName())

	if _, err := copyFile(testCaCertFile, filepath.Join(pkiOutputDir, config.GetCAPemFileName())); err != nil {
		return err
	}
	if _, err := copyFile(testCaKeyFile, filepath.Join(pkiOutputDir, config.GetCAPrivateKeyFileName())); err != nil {
		return err
	}
	if _, err := copyFile(testVaultCertFile, filepath.Join(pkiOutputDir, config.GetTLSPemFileName())); err != nil {
		return err
	}
	if _, err := copyFile(testVaultKeyFile, filepath.Join(pkiOutputDir, config.GetTLSPrivateKeyFileName())); err != nil {
		return err
	}

	return nil
}
