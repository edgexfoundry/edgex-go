//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"strconv"

	"github.com/spf13/cast"
)

type pair struct {
	Key   string
	Value string
}

func convertInterfaceToPairs(path string, interfaceMap any) []*pair {
	pairs := make([]*pair, 0)

	pathPre := ""
	if path != "" {
		pathPre = path + KeyDelimiter
	}

	switch value := interfaceMap.(type) {
	case []any:
		for index, item := range value {
			nextPairs := convertInterfaceToPairs(pathPre+strconv.Itoa(index), item)
			pairs = append(pairs, nextPairs...)
		}
	case map[string]any:
		for index, item := range value {
			nextPairs := convertInterfaceToPairs(pathPre+index, item)
			pairs = append(pairs, nextPairs...)
		}
	default:
		pairs = append(pairs, &pair{Key: path, Value: cast.ToString(value)})
	}

	return pairs
}
