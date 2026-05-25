//go:build non_delayedstart
// +build non_delayedstart

//
// Copyright (c) 2022 Intel Corporation
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

package runtimetokenprovider

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/types"
)

type runtimetokenprovider struct{}

func NewRuntimeTokenProvider(_ context.Context, _ logger.LoggingClient,
	_ types.RuntimeTokenProviderInfo) RuntimeTokenProvider {
	return &runtimetokenprovider{}
}

func (p *runtimetokenprovider) GetRawToken(serviceKey string) (string, error) {
	return "", fmt.Errorf("wrong build: RuntimeTokenProvider is not available. " +
		"Build without \"-tags non_delayedstart\" on the go build command line to enable runtime support for this feature.")
}
