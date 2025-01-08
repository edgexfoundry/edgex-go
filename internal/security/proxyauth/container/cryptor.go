//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils/crypto/interfaces"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

// CryptoInterfaceName contains the name of the interfaces.Crypto implementation in the DIC.
var CryptoInterfaceName = di.TypeInstanceToName((*interfaces.Crypto)(nil))

// CryptoFrom helper function queries the DIC and returns the interfaces.Cryptor implementation.
func CryptoFrom(get di.Get) interfaces.Crypto {
	return get(CryptoInterfaceName).(interfaces.Crypto)
}
