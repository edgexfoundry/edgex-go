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
	"io"

	"github.com/edgexfoundry/edgex-go/internal/pkg/utils/crypto/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

const aesKey = "RO6gGYKocUahpdX15k9gYvbLuSxbKrPz"

// AESCryptor defined the AES cryptor struct
type AESCryptor struct {
	key []byte
}

func NewAESCryptor() interfaces.Crypto {
	return &AESCryptor{
		key: []byte(aesKey),
	}
}

// Encrypt encrypts the given plaintext with AES-CBC mode and returns a string in base64 encoding
func (c *AESCryptor) Encrypt(plaintext string) (string, errors.EdgeX) {
	bytePlaintext := []byte(plaintext)
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "encrypt failed", err)
	}

	// CBC mode works on blocks so plaintexts may need to be padded to the next whole block
	paddedPlaintext := pkcs7Pad(bytePlaintext, block.BlockSize())

	ciphertext := make([]byte, aes.BlockSize+len(paddedPlaintext))
	// attach a random iv ahead of the ciphertext
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "encrypt failed", err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], paddedPlaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts the given ciphertext with AES-CBC mode and returns the original value as string
func (c *AESCryptor) Decrypt(ciphertext string) ([]byte, errors.EdgeX) {
	decodedCipherText, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed", err)
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed", err)
	}

	if len(decodedCipherText) < aes.BlockSize {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "decrypt failed", err)
	}

	// get the iv from the cipher text
	iv := decodedCipherText[:aes.BlockSize]
	decodedCipherText = decodedCipherText[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decodedCipherText, decodedCipherText)

	// If the original plaintext lengths are not a multiple of the block
	// size, padding would have to be added when encrypting, which would be
	// removed at this point
	plaintext, e := pkcs7Unpad(decodedCipherText)
	if e != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
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
