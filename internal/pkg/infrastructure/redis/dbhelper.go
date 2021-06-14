//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package redis

import "strings"

// CreateKey creates Redis key by connecting the target key with DBKeySeparator
func CreateKey(targets ...string) string {
	return strings.Join(targets, DBKeySeparator)
}
