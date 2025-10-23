//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
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
	aesSecretName = "aes"
	aesKeyName    = "key"

	keySourceSelf        = "self"
	keySourceSecretStore = "secret-store"
	keySourceLengthSize  = 4
	formatHeaderSize     = 4
	// gcmNonceSize defines the size of the nonce (number used once) for AES-GCM mode.
	// The nonce ensures that encrypting the same plaintext multiple times produces different ciphertexts.
	// 12 bytes is the standard size for optimal performance and security in GCM mode.
	gcmNonceSize = 12
	// gcmTagSize defines the size of the authentication tag for AES-GCM mode.
	// GCM produces a 16-byte authentication tag that provides integrity and authenticity verification.
	// This tag is automatically verified during decryption to detect any tampering or corruption.
	gcmTagSize            = 16
	minimumCipherTextSize = formatHeaderSize + keySourceLengthSize + gcmNonceSize + gcmTagSize
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

func NewAESCryptor(defaultKey []byte) interfaces.Crypto {
	return &AESCryptor{
		key:          defaultKey,
		keySource:    keySourceSelf,
		nonSecureKey: defaultKey,
	}
}

// NewAESCryptorWithSecretProvider creates a new AES cryptor that uses a key from SecretProvider
// If the key doesn't exist in the Secret Store, it generates a new one and stores it
func NewAESCryptorWithSecretProvider(secretProvider bootstrapInterfaces.SecretProvider, defaultKey []byte) (interfaces.Crypto, error) {
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
					nonSecureKey: defaultKey,
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
		nonSecureKey: defaultKey,
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

// Encrypt encrypts the given plaintext with AES-GCM mode
// ciphertext format: [formatHeader:4bytes][keySourceLength:4bytes][keySource:string][nonce:12bytes][encrypted_data]
func (c *AESCryptor) Encrypt(plaintext string) (string, errors.EdgeX) {
	bytePlaintext := []byte(plaintext)

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "encrypt failed", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "encrypt failed", err)
	}

	keySourceBytes := []byte(c.keySource)
	keySourceLength := uint32(len(keySourceBytes))

	// generate random nonce for GCM
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "encrypt failed", err)
	}

	// encrypt with GCM
	encryptedData := gcm.Seal(nil, nonce, bytePlaintext, nil)

	// calculate total size: formatHeader + keySourceLength + keySource + nonce + encryptedData
	totalSize := formatHeaderSize + keySourceLengthSize + len(keySourceBytes) + len(nonce) + len(encryptedData)
	ciphertext := make([]byte, totalSize)

	// write format header
	binary.BigEndian.PutUint32(ciphertext[:formatHeaderSize], newFormat)

	// write keySource length
	binary.BigEndian.PutUint32(ciphertext[formatHeaderSize:formatHeaderSize+keySourceLengthSize], keySourceLength)

	// write keySource string
	copy(ciphertext[formatHeaderSize+keySourceLengthSize:formatHeaderSize+keySourceLengthSize+len(keySourceBytes)], keySourceBytes)

	// write nonce
	nonceStart := formatHeaderSize + keySourceLengthSize + len(keySourceBytes)
	copy(ciphertext[nonceStart:nonceStart+len(nonce)], nonce)

	// write encrypted data
	copy(ciphertext[nonceStart+len(nonce):], encryptedData)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts the given ciphertext with AES-GCM mode for new format and AES-CBC mode for legacy format
// ciphertext format for AES-GCM: [formatHeader:4bytes][keySourceLength:4bytes][keySource:string][nonce:12bytes][encrypted_data]
// legacy format for AES-CBC: [iv:16bytes][encrypted_data]
func (c *AESCryptor) Decrypt(ciphertext string) ([]byte, errors.EdgeX) {
	decodedCipherText, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed", err)
	}

	if len(decodedCipherText) > formatHeaderSize {
		format := binary.BigEndian.Uint32(decodedCipherText[:formatHeaderSize])
		if format == newFormat {
			return c.decryptGCM(decodedCipherText)
		}
	}

	return c.decryptCBC(decodedCipherText)
}

// decryptGCM handles decryption for the GCM format
func (c *AESCryptor) decryptGCM(decodedCipherText []byte) ([]byte, errors.EdgeX) {
	if len(decodedCipherText) < minimumCipherTextSize {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed: ciphertext too short for GCM format", nil)
	}

	keySourceLength := binary.BigEndian.Uint32(decodedCipherText[formatHeaderSize : formatHeaderSize+keySourceLengthSize])
	keySourceBytes := decodedCipherText[formatHeaderSize+keySourceLengthSize : formatHeaderSize+keySourceLengthSize+keySourceLength]
	keySource := string(keySourceBytes)

	var keyToUse []byte
	switch keySource {
	case keySourceSelf:
		keyToUse = c.nonSecureKey
	case keySourceSecretStore:
		keyToUse = c.key
	default:
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed: unsupported keySource", nil)
	}

	block, err := aes.NewCipher(keyToUse)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed", err)
	}

	nonceStart := formatHeaderSize + keySourceLengthSize + keySourceLength
	nonce := decodedCipherText[nonceStart : nonceStart+gcmNonceSize]
	encryptedData := decodedCipherText[nonceStart+gcmNonceSize:]

	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed: GCM authentication failed", err)
	}

	return plaintext, nil
}

// decryptCBC handles decryption for the legacy CBC format
func (c *AESCryptor) decryptCBC(decodedCipherText []byte) ([]byte, errors.EdgeX) {
	if len(decodedCipherText) < aes.BlockSize {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed: ciphertext too short for CBC format", nil)
	}

	iv := decodedCipherText[:aes.BlockSize]
	encryptedText := decodedCipherText[aes.BlockSize:]

	block, err := aes.NewCipher(c.nonSecureKey)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed", err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decryptedText := make([]byte, len(encryptedText))
	mode.CryptBlocks(decryptedText, encryptedText)

	// remove PKCS7 padding
	plaintext, e := pkcs7Unpad(decryptedText)
	if e != nil {
		return nil, errors.NewCommonEdgeXWrapper(e)
	}

	return plaintext, nil
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
