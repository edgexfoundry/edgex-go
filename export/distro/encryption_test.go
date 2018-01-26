//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"

	"github.com/edgexfoundry/edgex-go/export"

	"testing"
)

const (
	plainString = "This is the test string used for testing"
	iv          = "123456789012345678901234567890"
	key         = "aquqweoruqwpeoruqwpoeruqwpoierupqoweiurpoqwiuerpqowieurqpowieurpoqiweuroipwqure"
)

func aesDecrypt(crypt []byte, aesData export.EncryptionDetails) []byte {
	hash := sha1.New()

	hash.Write([]byte((aesData.Key)))
	key := hash.Sum(nil)
	key = key[:blockSize]

	iv := make([]byte, blockSize)
	copy(iv, []byte(aesData.InitVector))

	block, err := aes.NewCipher(key)
	if err != nil {
		panic("key error")
	}

	decodedData, _ := base64.StdEncoding.DecodeString(string(crypt))

	ecb := cipher.NewCBCDecrypter(block, []byte(iv))
	decrypted := make([]byte, len(decodedData))
	ecb.CryptBlocks(decrypted, decodedData)

	trimmed := pkcs5Trimming(decrypted)

	return trimmed
}

func pkcs5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}

func TestAES(t *testing.T) {

	aesData := export.EncryptionDetails{
		Algo:       "AES",
		Key:        key,
		InitVector: iv,
	}

	enc := NewAESEncryption(aesData)

	cphrd := enc.Transform([]byte(plainString))

	decphrd := aesDecrypt(cphrd, aesData)

	if string(plainString) != string(decphrd) {
		t.Fatal("Encoded string ", string(plainString), " is not ", string(decphrd))
	}
}
