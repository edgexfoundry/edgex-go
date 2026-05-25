// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package algorithm

import (
	"fmt"
	"reflect"

	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaalgorithm"
)

type keyDerivationFactory struct{}

type KeyDerivationFactory interface {
	Hkdf(hashAlg HashAlgorithmType) *Algorithm
	TLS12PRF(hashAlg HashAlgorithmType) *Algorithm
	TLS12PSKToMs(hashAlg HashAlgorithmType) *Algorithm
}

func NewKeyDerivation() KeyDerivationFactory {
	return &keyDerivationFactory{}
}

func (f *keyDerivationFactory) Hkdf(hashAlg HashAlgorithmType) *Algorithm {
	return &Algorithm{
		variant: &KeyDerivation{
			variant: &KeyDerivationHkdf{
				HashAlg: hashAlg,
			},
		},
	}
}
func (f *keyDerivationFactory) TLS12PRF(hashAlg HashAlgorithmType) *Algorithm {
	return &Algorithm{
		variant: &KeyDerivation{
			variant: &KeyDerivationTLS12Prf{
				HashAlg: hashAlg,
			},
		},
	}
}
func (f *keyDerivationFactory) TLS12PSKToMs(hashAlg HashAlgorithmType) *Algorithm {
	return &Algorithm{
		variant: &KeyDerivation{
			variant: &KeyDerivationTLS12PskToMs{
				HashAlg: hashAlg,
			},
		},
	}
}

type KeyDerivation struct {
	//	*Algorithm_KeyDerivation_Hkdf_
	//	*Algorithm_KeyDerivation_Tls12Prf_
	//	*Algorithm_KeyDerivation_Tls12PskToMs_
	variant keyDerivationVariant
}

func (a *KeyDerivation) isAlgorithmVariant() {}
func (a *KeyDerivation) ToWireInterface() interface{} {
	return a.variant.ToWireInterface()
}

type keyDerivationVariant interface {
	toWire
	isKeyDerivationVariant()
	toWireInterfaceSpecific() *psaalgorithm.Algorithm_KeyDerivation
}

func (a *KeyDerivation) GetKeyDerivationHkdf() *KeyDerivationHkdf {
	if alg, ok := a.variant.(*KeyDerivationHkdf); ok {
		return alg
	}
	return nil
}

func (a *KeyDerivation) GetKeyDerivationTLS12Prf() *KeyDerivationTLS12Prf {
	if alg, ok := a.variant.(*KeyDerivationTLS12Prf); ok {
		return alg
	}
	return nil
}

func (a *KeyDerivation) GetKeyDerivationTLS12PskToMs() *KeyDerivationTLS12PskToMs {
	if alg, ok := a.variant.(*KeyDerivationTLS12PskToMs); ok {
		return alg
	}
	return nil
}

type KeyDerivationHkdf struct {
	HashAlg HashAlgorithmType
}

func (a *KeyDerivationHkdf) isKeyDerivationVariant() {}
func (a *KeyDerivationHkdf) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_KeyDerivation_{
			KeyDerivation: a.toWireInterfaceSpecific(),
		},
	}
}

func (a *KeyDerivationHkdf) toWireInterfaceSpecific() *psaalgorithm.Algorithm_KeyDerivation {
	return &psaalgorithm.Algorithm_KeyDerivation{
		Variant: &psaalgorithm.Algorithm_KeyDerivation_Hkdf_{
			Hkdf: &psaalgorithm.Algorithm_KeyDerivation_Hkdf{
				HashAlg: psaalgorithm.Algorithm_Hash(a.HashAlg),
			},
		},
	}
}

type KeyDerivationTLS12Prf struct {
	HashAlg HashAlgorithmType
}

func (a *KeyDerivationTLS12Prf) isKeyDerivationVariant() {}
func (a *KeyDerivationTLS12Prf) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_KeyDerivation_{
			KeyDerivation: a.toWireInterfaceSpecific(),
		},
	}
}
func (a *KeyDerivationTLS12Prf) toWireInterfaceSpecific() *psaalgorithm.Algorithm_KeyDerivation {
	return &psaalgorithm.Algorithm_KeyDerivation{
		Variant: &psaalgorithm.Algorithm_KeyDerivation_Tls12Prf_{
			Tls12Prf: &psaalgorithm.Algorithm_KeyDerivation_Tls12Prf{
				HashAlg: psaalgorithm.Algorithm_Hash(a.HashAlg),
			},
		},
	}
}

type KeyDerivationTLS12PskToMs struct {
	HashAlg HashAlgorithmType
}

func (a *KeyDerivationTLS12PskToMs) isKeyDerivationVariant() {}

func (a *KeyDerivationTLS12PskToMs) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_KeyDerivation_{
			KeyDerivation: a.toWireInterfaceSpecific(),
		},
	}
}
func (a *KeyDerivationTLS12PskToMs) toWireInterfaceSpecific() *psaalgorithm.Algorithm_KeyDerivation {
	return &psaalgorithm.Algorithm_KeyDerivation{
		Variant: &psaalgorithm.Algorithm_KeyDerivation_Tls12PskToMs_{
			Tls12PskToMs: &psaalgorithm.Algorithm_KeyDerivation_Tls12PskToMs{
				HashAlg: psaalgorithm.Algorithm_Hash(a.HashAlg),
			},
		},
	}
}

func newKeyDerivationFromWire(a *psaalgorithm.Algorithm_KeyDerivation) (*KeyDerivation, error) {
	switch linealg := a.Variant.(type) {
	case *psaalgorithm.Algorithm_KeyDerivation_Hkdf_:
		return &KeyDerivation{
			variant: &KeyDerivationHkdf{
				HashAlg: HashAlgorithmType(linealg.Hkdf.HashAlg),
			},
		}, nil
	case *psaalgorithm.Algorithm_KeyDerivation_Tls12Prf_:
		return &KeyDerivation{
			variant: &KeyDerivationTLS12Prf{
				HashAlg: HashAlgorithmType(linealg.Tls12Prf.HashAlg),
			},
		}, nil
	case *psaalgorithm.Algorithm_KeyDerivation_Tls12PskToMs_:
		return &KeyDerivation{
			variant: &KeyDerivationTLS12PskToMs{
				HashAlg: HashAlgorithmType(linealg.Tls12PskToMs.HashAlg),
			},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected type encountered decoding key derivation algorithm: %v", reflect.TypeOf(linealg))
	}
}
