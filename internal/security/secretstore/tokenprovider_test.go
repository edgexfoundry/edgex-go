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

package secretstore

import (
	"context"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/security/secretstoreclient"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/stretchr/testify/assert"
)

func TestInvalidProvider(t *testing.T) {
	config := secretstoreclient.SecretServiceInfo{
		TokenProvider:     "does-not-exist",
		TokenProviderType: OneShotProvider,
	}
	cancel, err := testCommon(config)
	defer cancel()
	assert.NotNil(t, err)
}

func TestInvalidProviderType(t *testing.T) {
	config := secretstoreclient.SecretServiceInfo{
		TokenProvider:     "/bin/true",
		TokenProviderType: "simple",
	}
	cancel, err := testCommon(config)
	defer cancel()
	assert.Error(t, err)
}

func TestNoConfig(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := NewTokenProvider(ctx, logger.MockLogger{})
	err := p.Launch()
	defer cancel()
	assert.Error(t, err)
}

func TestTrue(t *testing.T) {
	config := secretstoreclient.SecretServiceInfo{
		TokenProvider:     "/bin/true",
		TokenProviderType: OneShotProvider,
	}
	cancel, err := testCommon(config)
	defer cancel()
	assert.Nil(t, err)
}

func TestFalse(t *testing.T) {
	config := secretstoreclient.SecretServiceInfo{
		TokenProvider:     "/bin/false",
		TokenProviderType: OneShotProvider,
	}
	cancel, err := testCommon(config)
	defer cancel()
	assert.Error(t, err)
}

func TestEcho(t *testing.T) {
	config := secretstoreclient.SecretServiceInfo{
		TokenProvider:     "/bin/echo",
		TokenProviderArgs: []string{"one", "two"},
		TokenProviderType: OneShotProvider,
	}
	cancel, err := testCommon(config)
	defer cancel()
	assert.Nil(t, err)
}

func testCommon(config secretstoreclient.SecretServiceInfo) (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())
	p := NewTokenProvider(ctx, logger.MockLogger{})
	if err := p.SetConfiguration(config); err != nil {
		return cancel, err
	}
	return cancel, p.Launch()
}
