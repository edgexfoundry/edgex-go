//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package io

import (
	"io"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// To avoid large data causing unexpected memory exhaustion when decoding CBOR payload, defaultMaxEventSize was introduced as
// a reasonable limit appropriate for handling CBOR payload in edgex-go.  More details could be found at
// https://github.com/edgexfoundry/edgex-go/issues/2439
// TODO Make MaxEventSize a service configuration setting, so that users could adjust the limit per systems requirements
// https://github.com/edgexfoundry/edgex-go/issues/3237
const defaultMaxEventSize = int64(25 * 1e6) // 25 MB

func ReadAddEventRequestInBytes(reader io.Reader) ([]byte, errors.EdgeX) {
	// use LimitReader with defaultMaxEventSize to avoid unexpected memory exhaustion
	bytes, err := io.ReadAll(io.LimitReader(reader, defaultMaxEventSize))
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindIOError, "AddEventRequest I/O reading failed", err)
	}
	return bytes, nil
}
