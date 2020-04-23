/*******************************************************************************
 * Copyright 2020 Redis Labs
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

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/edgexfoundry/edgex-go/internal/security/secretread"
)

func main() {
	lc := logger.NewClient(secretread.ServiceKey, false, "", models.ErrorLog)

	config, err := secretread.LoadConfig(lc)
	if err != nil {
		lc.Error(fmt.Sprintf("Failed to load configration: %s", err.Error()))
		os.Exit(1)
	}

	databases, err := secretread.GetCredentials(lc, config)
	if err != nil {
		lc.Error((fmt.Sprintf("Failed to get credentials: %s", err.Error())))
		os.Exit(1)
	}

	dir := filepath.Dir(config.SecretStore.PasswordFile)
	if err := os.MkdirAll(dir, os.FileMode(0700)); err != nil {
		lc.Error((fmt.Sprintf("Failed to create directory %s: %s", dir, err.Error())))
		os.Exit(1)
	}

	if err := ioutil.WriteFile(config.SecretStore.PasswordFile, []byte(databases["redis5"].Password), os.FileMode(0600)); err != nil {
		lc.Error((fmt.Sprintf("Failed to create file %s: %s", config.SecretStore.PasswordFile, err.Error())))
		os.Exit(1)
	}

	os.Exit(0)
}
