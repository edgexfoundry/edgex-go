//
// Copyright (c) 2017
// Mainflux
// Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package export

// Encryption types
const (
	EncNone = "NONE"
	EncAes  = "AES"
)

// EncryptionDetails - Provides details for encryption
// of export data per client request
type EncryptionDetails struct {
	Algo       string `bson:"encryptionAlgorithm,omitempty" json:"encryptionAlgorithm,omitempty"`
	Key        string `bson:"encryptionKey,omitempty" json:"encryptionKey,omitempty"`
	InitVector string `bson:"initializingVector,omitempty" json:"initializingVector,omitempty"`
}
