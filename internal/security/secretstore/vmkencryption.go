//
// Copyright (c) 2021 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//
// US Export Control Classification Number (ECCN): 5D002TSU
//

package secretstore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/security/kdf"
	"github.com/edgexfoundry/edgex-go/internal/security/pipedhexreader"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"
)

/*

THE FOLLOWING DISCLAIMER IS REQUIRED TO BE CARRIED WITH THE BELOW CODE
DO NOT REMOVE EXCEPT BY PERMISSION OF THE AUTHOR

Vault Master Key Encryption Feature

The purpose of this feature is to provide a secure way to unlock Vault
without requiring human intervention to supply the Vault Master Key.

This feature requires a sufficiently secure source of randomness in order
to unlock the Vault. This randomness should have at least 256 bits of entropy
and be backed by hardware secure storage.
If using a TPM, the secret should be bound to platform
configuration registers to attest the system state.

The randomness should be output as a string of hex-encoded octets to standard
output and the executable (or path to the executable) should be specified in the
IMK_HOOK environment variable.

In addition, the underlying platform that serves as the execution platform for
EdgeX must be secured against the ability for an attacker to intercept the
source of randomness between retrieving it out of hardware secure storage and
passing it to security-secretstore-setup. This may entail use of secure boot,
dm-verity protected partitions, volume encryption, orchestrator hardening,
and other measures.

*/

const aesKeyLength = 32 // for AES-256

type VMKEncryption struct {
	fileOpener     fileioperformer.FileIoPerformer
	pipedHexReader pipedhexreader.PipedHexReader
	kdf            kdf.KeyDeriver
	encrypting     bool
	ikm            []byte
}

// NewVMKEncryption - constructor
func NewVMKEncryption(fileOpener fileioperformer.FileIoPerformer,
	pipedHexReader pipedhexreader.PipedHexReader,
	kdf kdf.KeyDeriver) *VMKEncryption {

	return &VMKEncryption{
		fileOpener:     fileOpener,
		pipedHexReader: pipedHexReader,
		kdf:            kdf,
		encrypting:     false,
	}
}

// LoadIKM loads input key material from the specified path
func (v *VMKEncryption) LoadIKM(ikmBinPath string) error {
	if ikmBinPath == "" {
		return fmt.Errorf("ikmBinPath is required")
	}
	ikm, err := v.pipedHexReader.ReadHexBytesFromExe(ikmBinPath)
	if err != nil {
		return fmt.Errorf("Error reading input key material from IKM_HOOK - encryption not enabled: %w", err)
	}
	v.ikm = ikm
	v.encrypting = true
	return nil
}

// WipeIKM scrubs the input key material from memory
func (v *VMKEncryption) WipeIKM() {
	// Note: make() is defined to zero-fill the array
	copy(v.ikm, make([]byte, len(v.ikm)))
	v.encrypting = false
}

// IsEncrypting scrubs the input key material from memory
func (v *VMKEncryption) IsEncrypting() bool {
	return v.encrypting
}

// EncryptInitResponse processes the InitResponse and encrypts the key shares
// in the end, Keys and KeysBase64 are removed and replaced with
// EncryptedKeys and Nonces in the resulting JSON
// Root token is left untouched
func (v *VMKEncryption) EncryptInitResponse(initResp *types.InitResponse) error {
	// Check prerequisite (key has been loaded)
	if !v.encrypting {
		return fmt.Errorf("cannot encrypt init response as key has not been loaded")
	}

	newKeys := make([]string, len(initResp.Keys))
	newNonces := make([]string, len(initResp.Keys))

	for i, hexPlaintext := range initResp.Keys {

		plainText, err := hex.DecodeString(hexPlaintext)
		if err != nil {
			return fmt.Errorf("failed to decode hex bytes of keyshare (details omitted): %w", err)
		}

		keyShare, nonce, err := v.gcmEncryptKeyShare(plainText, i) // Wrap using a unique AES key
		if err != nil {
			return fmt.Errorf("failed to wrap key %d: %w", i, err)
		}

		newKeys[i] = hex.EncodeToString(keyShare)
		newNonces[i] = hex.EncodeToString(nonce)

		wipeKey(keyShare) // Clear out binary version of encrypted key
		wipeKey(nonce)    // Clear out nonce
	}

	initResp.EncryptedKeys = newKeys
	initResp.Nonces = newNonces
	initResp.Keys = nil       // strings are immutable, must wait for GC
	initResp.KeysBase64 = nil // strings are immutable, must wait for GC
	return nil
}

// DecryptInitResponse processes the InitResponse and decrypts the key shares
// in the end, EncryptedKeys and Nonces are removed and replaced with
// Keys and KeysBase64 in the resulting JSON like the init response was originally
// Root token is left untouched
func (v *VMKEncryption) DecryptInitResponse(initResp *types.InitResponse) error {
	// Check prerequisite (key has been loaded)
	if !v.encrypting {
		return fmt.Errorf("cannot decrypt init response as key has not been loaded")
	}

	newKeys := make([]string, len(initResp.EncryptedKeys))
	newKeysBase64 := make([]string, len(initResp.EncryptedKeys))

	for i, hexCiphertext := range initResp.EncryptedKeys {
		hexNonce := initResp.Nonces[i]
		nonce, err := hex.DecodeString(hexNonce)
		if err != nil {
			return fmt.Errorf("failed to decode hex bytes of nonce: %w", err)
		}

		cipherText, err := hex.DecodeString(hexCiphertext)
		if err != nil {
			return fmt.Errorf("failed to decode hex bytes of ciphertext: %w", err)
		}

		keyShare, err := v.gcmDecryptKeyShare(cipherText, nonce, i) // Unwrap using a unique AES key
		if err != nil {
			return fmt.Errorf("failed to unwrap key %d: %w", i, err)
		}

		newKeys[i] = hex.EncodeToString(keyShare)
		newKeysBase64[i] = base64.StdEncoding.EncodeToString(keyShare)
	}

	initResp.Keys = newKeys
	initResp.KeysBase64 = newKeysBase64
	initResp.EncryptedKeys = nil
	initResp.Nonces = nil

	return nil
}

//
// Internal methods
//

// gcmEncryptKeyShare encrypts each key share with a unique key
// from the key derivation function based on passing the info
// string vault0, vault1, ... et cetera to the KDF.
func (v *VMKEncryption) gcmEncryptKeyShare(keyShare []byte, counter int) ([]byte, []byte, error) {

	defer wipeKey(keyShare) // wipe original keyShare on exit

	info := fmt.Sprintf("vault%d", counter)

	key, err := v.kdf.DeriveKey(v.ikm, aesKeyLength, info)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive encryption key for vault master key share %w", err)
	}
	defer wipeKey(key) // wipe encryption key on exit

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize block cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize AES cipher: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize random nonce: %w", err)
	}

	// Encrypt the key share (plaintext to be wiped on exit by deferred function)
	ciphertext := aesGCM.Seal(nil, nonce, keyShare, nil)

	return ciphertext, nonce, nil
}

// gcmDecryptKeyShare decrypts each key share with a unique key
// from the key derivation function based on passing the info
// string vault0, vault1, ... et cetera to the KDF.
func (v *VMKEncryption) gcmDecryptKeyShare(keyShare []byte, nonce []byte, counter int) ([]byte, error) {

	defer wipeKey(keyShare) // wipe original (encrypted) key share on exit (not technically needed)

	info := fmt.Sprintf("vault%d", counter)

	key, err := v.kdf.DeriveKey(v.ikm, aesKeyLength, info)
	if err != nil {
		return nil, fmt.Errorf("failed to derive encryption key for vault master key share %w", err)
	}
	defer wipeKey(key) // wipe encryption key on exit

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize block cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AES cipher: %w", err)
	}

	// Decrypt key share; on error, erase any partial results
	plaintext, err := aesGCM.Open(nil, nonce, keyShare, nil)
	if err != nil {
		if plaintext != nil {
			wipeKey(plaintext)
		}
		return nil, fmt.Errorf("failed to decrypt ciphertext: %w", err)
	}

	return plaintext, nil
}

func wipeKey(key []byte) {
	blank := make([]byte, len(key)) // zero-filled
	copy(key, blank)
}
