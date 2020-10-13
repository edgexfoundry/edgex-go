//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"time"
)

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// FindCommonStrings finds the common string from multiple string slices
// e.g.
// stringSlice1 = []string{"aaa", "bbb", "ccc"}
// stringSlice2 = []string{"bbb", "ccc", "ddd"}
// stringSlice3 = []string{"aaa", "bbb", "ccc", "eee"}
// FindCommonStrings(stringSlice1, stringSlice2, stringSlice3) shall return a string slice with {"bbb", "ccc"}
func FindCommonStrings(stringsArray ...[]string) (commonStrings []string) {
	if len(stringsArray) == 0 {
		return nil
	} else if len(stringsArray) == 1 {
		return stringsArray[0]
	}
	commonStringsSet := make([]string, 0)
	hash := make(map[string]bool)
	for _, s := range stringsArray[0] {
		hash[s] = true
	}
	for _, s := range stringsArray[1] {
		if _, ok := hash[s]; ok {
			commonStringsSet = append(commonStringsSet, s)
		}
	}
	stringsArray = append([][]string{commonStringsSet}, stringsArray[2:]...)
	return FindCommonStrings(stringsArray...)
}

// ConvertStringsToInterfaces converts a string array to an interface{} array
func ConvertStringsToInterfaces(stringArray []string) []interface{} {
	result := make([]interface{}, len(stringArray))
	for i, s := range stringArray {
		result[i] = s
	}
	return result
}
