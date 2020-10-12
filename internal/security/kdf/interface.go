//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package kdf

// KeyDeriver is the interface that the main program expects
// for returning a derived key.
type KeyDeriver interface {
	// DeriveKey returns a byte array that is of keyLen length and
	// an error if errors where encountered while deriving the key
	// inputKeyingMaterial and info are inputs
	// to the key deriviation function, an keyLen
	// is the desired length of the derived key.
	// inputKeyingMaterial is a secret
	// and info is used to cause the KDF to generate
	// different output keys from the same inputKeyingMaterial.
	// Please see the application notes for RFC 5869
	// https://tools.ietf.org/html/rfc5869#3 for
	// details for details about the key derivation algorithm.
	DeriveKey(inputKeyingMaterial []byte, keyLen uint, info string) ([]byte, error)
}
