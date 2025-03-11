// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package algorithm

import (
	"fmt"
	"reflect"

	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaalgorithm"
)

type AsymmetricEncryptionFactory interface {
	RsaPkcs1V15Crypt() *Algorithm
	RsaOaep(hashAlg HashAlgorithmType) *Algorithm
}

type asymmetricEncryptionFactory struct{}

func NewAsymmetricEncryption() AsymmetricEncryptionFactory {
	return &asymmetricEncryptionFactory{}
}

func (a *asymmetricEncryptionFactory) RsaPkcs1V15Crypt() *Algorithm {
	return &Algorithm{
		variant: &AsymmetricEncryptionAlgorithm{
			variant: &AsymmetricEncryptionRsaPkcs1V15Crypt{},
		},
	}
}

func (a *asymmetricEncryptionFactory) RsaOaep(hashAlg HashAlgorithmType) *Algorithm {
	return &Algorithm{
		variant: &AsymmetricEncryptionAlgorithm{
			variant: &AsymmetricEncryptionRsaOaep{
				HashAlg: hashAlg,
			},
		},
	}
}

type AsymmetricEncryptionAlgorithm struct {
	//	*AsymmetricEncryptionRsaPkcs1V15Crypt
	//	*AsymmetricEncryptionRsaOaep
	variant asymmetricEncryptionAlgorithmVariant
}

type asymmetricEncryptionAlgorithmVariant interface {
	toWire
	// Algorithm
	isAsymmetricEncryptionAlgorithmVariant()
}

func (a *AsymmetricEncryptionAlgorithm) ToWireInterface() interface{} {
	return a.variant.ToWireInterface()
}

func (a *AsymmetricEncryptionAlgorithm) isAlgorithmVariant() {}

func (a *AsymmetricEncryptionAlgorithm) GetRsaPkcs1V15Crypt() *AsymmetricEncryptionRsaPkcs1V15Crypt {
	if alg, ok := a.variant.(*AsymmetricEncryptionRsaPkcs1V15Crypt); ok {
		return alg
	}
	return nil
}

func (a *AsymmetricEncryptionAlgorithm) GetRsaOaep() *AsymmetricEncryptionRsaOaep {
	if alg, ok := a.variant.(*AsymmetricEncryptionRsaOaep); ok {
		return alg
	}
	return nil
}

func newAsymmetricEncryptionFromWire(a *psaalgorithm.Algorithm_AsymmetricEncryption) (*AsymmetricEncryptionAlgorithm, error) {
	if a == nil || a.Variant == nil {
		return nil, fmt.Errorf("nil argument passed")
	}
	switch alg := a.Variant.(type) {
	case *psaalgorithm.Algorithm_AsymmetricEncryption_RsaPkcs1V15Crypt_:
		return NewAsymmetricEncryption().RsaPkcs1V15Crypt().GetAsymmetricEncryption(), nil
	case *psaalgorithm.Algorithm_AsymmetricEncryption_RsaOaep_:
		return NewAsymmetricEncryption().RsaOaep(HashAlgorithmType(alg.RsaOaep.HashAlg)).GetAsymmetricEncryption(), nil
	default:
		return nil, fmt.Errorf("expected *AsymmetricEncryption compatible type, got %v", reflect.TypeOf(alg))
	}
}

type AsymmetricEncryptionRsaPkcs1V15Crypt struct {
}

func (a *AsymmetricEncryptionRsaPkcs1V15Crypt) isAsymmetricEncryptionAlgorithmVariant() {}

func (a *AsymmetricEncryptionRsaPkcs1V15Crypt) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_AsymmetricEncryption_{
			AsymmetricEncryption: &psaalgorithm.Algorithm_AsymmetricEncryption{
				Variant: &psaalgorithm.Algorithm_AsymmetricEncryption_RsaPkcs1V15Crypt_{},
			},
		},
	}
}

type AsymmetricEncryptionRsaOaep struct {
	HashAlg HashAlgorithmType
}

func (a *AsymmetricEncryptionRsaOaep) isAsymmetricEncryptionAlgorithmVariant() {}

func (a *AsymmetricEncryptionRsaOaep) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_AsymmetricEncryption_{
			AsymmetricEncryption: &psaalgorithm.Algorithm_AsymmetricEncryption{
				Variant: &psaalgorithm.Algorithm_AsymmetricEncryption_RsaOaep_{
					RsaOaep: &psaalgorithm.Algorithm_AsymmetricEncryption_RsaOaep{
						HashAlg: psaalgorithm.Algorithm_Hash(a.HashAlg),
					},
				},
			},
		},
	}
}
