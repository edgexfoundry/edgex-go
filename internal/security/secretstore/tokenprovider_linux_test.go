//go:build linux
// +build linux

//
// Copyright (c) 2021 Intel Corporation
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
// SPDX-License-Identifier: Apache-2.0
//

package secretstore

import (
	"context"
	"os"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstore/config"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"

	"github.com/stretchr/testify/assert"
)

// TestCreatesFile only runs on Linux and makes sure the code can
// run a real executable taking real arguments.
func TestCreatesFile(t *testing.T) {
	const testfile = "/tmp/tokenprovider_linux_test.dat"
	configuration := config.SecretStoreInfo{
		TokenProvider:     "touch",
		TokenProviderType: OneShotProvider,
		TokenProviderArgs: []string{testfile},
	}
	ctx, cancel := context.WithCancel(context.Background())

	err := os.RemoveAll(testfile)
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(testfile) }() // cleanup

	p := NewTokenProvider(ctx, logger.MockLogger{}, NewDefaultExecRunner())
	err = p.SetConfiguration(configuration)
	assert.NoError(t, err)

	err = p.Launch()
	assert.NoError(t, err)
	defer cancel()

	file, err := os.Open(testfile)
	assert.NoError(t, err) // fails if file wasn't created
	defer file.Close()
}
