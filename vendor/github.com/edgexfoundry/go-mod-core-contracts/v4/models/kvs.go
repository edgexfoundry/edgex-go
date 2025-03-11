//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

// KVResponse defines the Response Content for GET Keys of core-keeper (GET /kvs/key/{key} API).
type KVResponse interface {
	SetKey(string)
}

// KVS defines the Response Content for GET Keys controller with keyOnly is false which inherits KVResponse interface.
type KVS struct {
	Key string `json:"key,omitempty"`
	StoredData
}

type StoredData struct {
	DBTimestamp
	Value interface{} `json:"value,omitempty"`
}

// KeyOnly defines the Response Content for GET Keys  with keyOnly is true which inherits KVResponse interface.
type KeyOnly string

func (kv *KVS) SetKey(newKey string) {
	kv.Key = newKey
}

func (key *KeyOnly) SetKey(newKey string) {
	*key = KeyOnly(newKey)
}
