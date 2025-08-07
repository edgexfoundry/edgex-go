//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/edgexfoundry/edgex-go/internal/pkg/utils/crypto/interfaces"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg"
)

const (
	aesKey        = "RO6gGYKocUahpdX15k9gYvbLuSxbKrPz"
	aesSecretName = "aes"
	aesKeyName    = "key"

	keySourceSelf        = "self"
	keySourceSecretStore = "secret-store"
	keySourceLengthSize  = 4
	formatHeaderSize     = 4
	// newFormat is a magic number to indicate the new format of the ciphertext.
	// The probability of a random legacy ciphertext starting with the same 4 bytes is 1 in 2^32.
	newFormat = 0x107ECA18
)

// AESCryptor defined the AES cryptor struct
type AESCryptor struct {
	key          []byte
	keySource    string
	nonSecureKey []byte
}

func NewAESCryptor() interfaces.Crypto {
	return &AESCryptor{
		key:          []byte(aesKey),
		keySource:    keySourceSelf,
		nonSecureKey: []byte(aesKey),
	}
}

// NewAESCryptorWithSecretProvider creates a new AES cryptor that uses a key from SecretProvider
// If the key doesn't exist in the Secret Store, it generates a new one and stores it
func NewAESCryptorWithSecretProvider(secretProvider bootstrapInterfaces.SecretProvider) (interfaces.Crypto, error) {
	if secretProvider == nil {
		return nil, fmt.Errorf("secret provider is nil, cannot create AESCryptor")
	}

	secrets, err := secretProvider.GetSecret(aesSecretName)
	if err == nil {
		if aesKeyStr, ok := secrets[aesKeyName]; ok && aesKeyStr != "" {
			keyBytes, decodeErr := base64.StdEncoding.DecodeString(aesKeyStr)
			if decodeErr == nil {
				return &AESCryptor{
					key:          keyBytes,
					keySource:    keySourceSecretStore,
					nonSecureKey: []byte(aesKey),
				}, nil
			}
			return nil, fmt.Errorf("invalid AES key format in secret store: %w", decodeErr)
		}
	} else if _, ok := err.(pkg.ErrSecretNameNotFound); !ok {
		return nil, fmt.Errorf("failed to get AES key from secret store: %w", err)
	}

	keyBytes, err := generateAndStoreNewAESKey(secretProvider)
	if err != nil {
		return nil, err
	}

	return &AESCryptor{
		key:          keyBytes,
		keySource:    keySourceSecretStore,
		nonSecureKey: []byte(aesKey),
	}, nil
}

func generateAndStoreNewAESKey(secretProvider bootstrapInterfaces.SecretProvider) ([]byte, error) {
	newKey := make([]byte, 32) // 256-bit (32 bytes) AES key
	if _, err := io.ReadFull(rand.Reader, newKey); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	keyBase64 := base64.StdEncoding.EncodeToString(newKey)
	secrets := map[string]string{
		aesKeyName: keyBase64,
	}

	if err := secretProvider.StoreSecret(aesSecretName, secrets); err != nil {
		return nil, fmt.Errorf("failed to store AES key in secret store: %w", err)
	}

	return newKey, nil
}

// Encrypt encrypts the given plaintext with AES-CBC mode and returns a string in base64 encoding
// The ciphertext format is: [formatHeader:4bytes][keySourceLength:4bytes][keySource:string][iv:16bytes][encrypted_data]
func (c *AESCryptor) Encrypt(plaintext string) (string, errors.EdgeX) {
	bytePlaintext := []byte(plaintext)
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "encrypt failed", err)
	}

	// AES encryption in CBC (Cipher Block Chaining) mode processes data in blocks of a fixed size (e.g., 32 bytes for AES).
	// If the plaintext length is not a multiple of the block size, the encryption process would fail.
	// Padding ensures that the plaintext is extended to the required length without altering its original content.
	paddedPlaintext := pkcs7Pad(bytePlaintext, block.BlockSize())

	keySourceBytes := []byte(c.keySource)
	keySourceLength := uint32(len(keySourceBytes))

	ciphertext := make([]byte, formatHeaderSize+keySourceLengthSize+len(keySourceBytes)+aes.BlockSize+len(paddedPlaintext))

	// write format header at the beginning
	binary.BigEndian.PutUint32(ciphertext[:formatHeaderSize], newFormat)

	// write keySource length
	binary.BigEndian.PutUint32(ciphertext[formatHeaderSize:formatHeaderSize+keySourceLengthSize], keySourceLength)

	// write keySource string
	copy(ciphertext[formatHeaderSize+keySourceLengthSize:formatHeaderSize+keySourceLengthSize+len(keySourceBytes)], keySourceBytes)

	// generate random IV (Initialization Vector)
	// the IV is a random value that ensures the same plaintext encrypted multiple times will yield different ciphertexts
	iv := ciphertext[formatHeaderSize+keySourceLengthSize+len(keySourceBytes) : formatHeaderSize+keySourceLengthSize+len(keySourceBytes)+aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "encrypt failed", err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[formatHeaderSize+keySourceLengthSize+len(keySourceBytes)+aes.BlockSize:], paddedPlaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts the given ciphertext with AES-CBC mode and returns the original value as string
// Supports both old format: [iv:16bytes][encrypted_data] and new format: [formatHeader:4bytes][keySourceLength:4bytes][keySource:string][iv:16bytes][encrypted_data]
func (c *AESCryptor) Decrypt(ciphertext string) ([]byte, errors.EdgeX) {
	decodedCipherText, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed", err)
	}

	if len(decodedCipherText) < aes.BlockSize {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed: ciphertext too short", nil)
	}

	var keySource string
	var iv []byte
	var encryptedText []byte
	keyToUse := c.nonSecureKey // default to non-secure key

	format := binary.BigEndian.Uint32(decodedCipherText[:formatHeaderSize])
	if format == newFormat {
		keySourceLength := binary.BigEndian.Uint32(decodedCipherText[formatHeaderSize : formatHeaderSize+keySourceLengthSize])
		keySourceBytes := decodedCipherText[formatHeaderSize+keySourceLengthSize : formatHeaderSize+keySourceLengthSize+keySourceLength]
		keySource = string(keySourceBytes)
		switch keySource {
		case keySourceSelf:
		case keySourceSecretStore:
			keyToUse = c.key
		default:
			return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed: unsupported keySource", nil)
		}

		ivStart := formatHeaderSize + keySourceLengthSize + keySourceLength
		iv = decodedCipherText[ivStart : ivStart+aes.BlockSize]
		encryptedText = decodedCipherText[ivStart+aes.BlockSize:]
	} else {
		iv = decodedCipherText[:aes.BlockSize]
		encryptedText = decodedCipherText[aes.BlockSize:]
	}

	block, err := aes.NewCipher(keyToUse)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed", err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decryptedText := make([]byte, len(encryptedText))
	mode.CryptBlocks(decryptedText, encryptedText)

	// If the original plaintext lengths are not a multiple of the block
	// size, padding would have to be added when encrypting, which would be
	// removed at this point
	plaintext, e := pkcs7Unpad(decryptedText)
	if e != nil {
		return nil, errors.NewCommonEdgeXWrapper(e)
	}

	return plaintext, nil
}

// pkcs7Pad implements the PKCS7 padding
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7Unpad implements the PKCS7 unpadding
func pkcs7Unpad(data []byte) ([]byte, errors.EdgeX) {
	length := len(data)
	unpadding := int(data[length-1])
	if unpadding > length {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "invalid padding", nil)
	}
	return data[:(length - unpadding)], nil
}
