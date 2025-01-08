//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import "github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

type DBClient interface {
	AddKey(name string, content string) errors.EdgeX
	UpdateKey(name string, content string) errors.EdgeX
	ReadKeyContent(name string) (string, errors.EdgeX)
	KeyExists(name string) (bool, errors.EdgeX)
}
