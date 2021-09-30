//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//
// US Export Control Classification Number (ECCN): 5D002TSU
//

// Package kdf implements the key deriviation function (KDF)
// for creation of encryption keys to protect the Vault key shares
package kdf

import (
	"crypto/rand"
	"errors"
	"hash"
	"os"
	"path"

	"golang.org/x/crypto/hkdf"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer"
)

// The KDF salt adds additional randomness to the KDF input key
// and provides a mechanism to force generation of unique keys
// in the event that the KDF inputKeyMaterial is less than random.
const (
	saltFile   string = "kdf-salt.dat"
	saltLength int    = 32
)

var (
	osStat = os.Stat
)

// kdfObject stores instance data for the default KDF
type kdfObject struct {
	fileIoPerformer fileioperformer.FileIoPerformer
	persistencePath string
	hashConstructor func() hash.Hash
}

// NewKdf creates a new KeyDeriver
func NewKdf(fileIoPerformer fileioperformer.FileIoPerformer, persistencePath string, hashConstructor func() hash.Hash) KeyDeriver {
	return &kdfObject{fileIoPerformer, persistencePath, hashConstructor}
}

// DeriveKey returns derived key material of specified length
func (kdf *kdfObject) DeriveKey(inputKeyingMaterial []byte, keyLen uint, info string) ([]byte, error) {
	salt, err := kdf.initializeSalt()
	if err != nil {
		return nil, err
	}
	infoBytes := []byte(info)
	kdfReader := hkdf.New(kdf.hashConstructor, inputKeyingMaterial, salt, infoBytes)
	key := make([]byte, keyLen)
	if _, err := kdfReader.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// initializeSalt recovers the KDF salt value from a file
// or installs a new salt
func (kdf *kdfObject) initializeSalt() ([]byte, error) {
	salt := make([]byte, saltLength)
	saltPath := path.Join(kdf.persistencePath, saltFile)

	_, err := osStat(saltPath)
	if err == nil {
		// File exists; read out the salt
		saltFileReader, err := kdf.fileIoPerformer.OpenFileReader(saltPath, os.O_RDONLY, 0400)
		if err != nil {
			return nil, err
		}

		saltFileObj := fileioperformer.MakeReadCloser(saltFileReader)
		defer saltFileObj.Close() // defer close for reading

		nbytes, err := saltFileObj.Read(salt)
		if err != nil {
			return nil, err
		}
		if nbytes != saltLength {
			return nil, errors.New("Salt file does not contain expected length of salt")
		}

		return salt, nil
	}

	// Check error code from os.Stat and if file does not exist, create the Salt
	if os.IsNotExist(err) {
		_, err := rand.Read(salt) // all of salt guaranteed to be filled if err==nil
		if err != nil {
			return nil, err
		}

		// os.O_TRUNC necessary to prevent TOCTOU issues if something wrote the file between the above stat and the creat() here
		saltFileWriter, err := kdf.fileIoPerformer.OpenFileWriter(saltPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return nil, err
		}
		saltFileObj := saltFileWriter
		// use explicit close() for writing
		if saltFileObj == nil {
			_ = saltFileWriter.Close()
			return nil, errors.New("saltFileWriter does not implement required read/write methods")
		}

		nwritten, err := saltFileObj.Write(salt)
		closeErr := saltFileObj.Close()
		if err != nil {
			return nil, err
		}
		if nwritten != len(salt) {
			err := errors.New("Failed to write entire contents of salt file; encryption key will likely be unrecoverable")
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
		return salt, nil
	}

	// Unexpected error from os.Stat()
	return nil, err
}
