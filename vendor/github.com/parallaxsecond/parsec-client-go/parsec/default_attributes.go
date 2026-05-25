// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package parsec

import "github.com/parallaxsecond/parsec-client-go/parsec/algorithm"

type DefaultKeyAttributeFactory interface {
	SigningKey() *KeyAttributes
}

type defaultKeyAttributeFactory struct{}

func DefaultKeyAttribute() DefaultKeyAttributeFactory {
	return &defaultKeyAttributeFactory{}
}

func (f *defaultKeyAttributeFactory) SigningKey() *KeyAttributes {
	const keyBits = 2048
	const hashAlg = algorithm.HashAlgorithmTypeSHA256
	return &KeyAttributes{
		KeyBits: keyBits,
		KeyType: NewKeyType().RsaKeyPair(),
		KeyPolicy: &KeyPolicy{
			KeyAlgorithm: algorithm.NewAsymmetricSignature().RsaPkcs1V15Sign(hashAlg),
			KeyUsageFlags: &UsageFlags{
				Cache:         false,
				Copy:          false,
				Decrypt:       false,
				Derive:        false,
				Encrypt:       false,
				Export:        false,
				SignHash:      true,
				SignMessage:   true,
				VerifyHash:    true,
				VerifyMessage: true,
			},
		},
	}
}
