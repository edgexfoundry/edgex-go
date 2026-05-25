// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package algorithm

import (
	"fmt"
	"reflect"

	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaalgorithm"
)

type KeyAgreementFactory interface {
	RawFFDH() *Algorithm
	RawECDH() *Algorithm
	FFDH(*KeyDerivation) *Algorithm
	ECDH(*KeyDerivation) *Algorithm
}

func NewKeyAgreement() KeyAgreementFactory {
	return &keyAgreementFactory{}
}

type keyAgreementFactory struct{}

func (f *keyAgreementFactory) RawFFDH() *Algorithm {
	return &Algorithm{
		variant: &KeyAgreement{
			variant: &KeyAgreementRaw{
				RawAlg: KeyAgreementFFDH,
			},
		},
	}
}
func (f *keyAgreementFactory) RawECDH() *Algorithm {
	return &Algorithm{
		variant: &KeyAgreement{
			variant: &KeyAgreementRaw{
				RawAlg: KeyAgreementECDH,
			},
		},
	}
}
func (f *keyAgreementFactory) FFDH(kd *KeyDerivation) *Algorithm {
	return &Algorithm{
		variant: &KeyAgreement{
			variant: &KeyAgreementWithKeyDerivation{
				DerivationAlg: kd,
				KaAlg:         KeyAgreementFFDH,
			},
		},
	}
}
func (f *keyAgreementFactory) ECDH(kd *KeyDerivation) *Algorithm {
	return &Algorithm{
		variant: &KeyAgreement{
			variant: &KeyAgreementWithKeyDerivation{
				DerivationAlg: kd,
				KaAlg:         KeyAgreementECDH,
			},
		},
	}
}

type KeyAgreementRawType int32

const (
	KeyAgreementRAWNONE KeyAgreementRawType = 0 // This default variant should not be used.
	KeyAgreementFFDH    KeyAgreementRawType = 1
	KeyAgreementECDH    KeyAgreementRawType = 2
)

func newKeyAgreementFromWire(a *psaalgorithm.Algorithm_KeyAgreement) (*KeyAgreement, error) {
	switch linealg := a.Variant.(type) {
	case *psaalgorithm.Algorithm_KeyAgreement_Raw_:
		return &KeyAgreement{
			variant: &KeyAgreementRaw{
				RawAlg: KeyAgreementRawType(linealg.Raw),
			},
		}, nil
	case *psaalgorithm.Algorithm_KeyAgreement_WithKeyDerivation_:
		kdf, err := newKeyDerivationFromWire(linealg.WithKeyDerivation.KdfAlg)
		if err != nil {
			return nil, err
		}
		return &KeyAgreement{
			variant: &KeyAgreementWithKeyDerivation{
				KaAlg:         KeyAgreementRawType(linealg.WithKeyDerivation.KaAlg),
				DerivationAlg: kdf,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected type when decoding key agreement algorithm: %v", reflect.TypeOf(linealg))
	}
}

type KeyAgreement struct {
	//	*KeyAgreementRaw
	//	*KeyAgreementWithKeyDerivation
	variant keyAgreementVariant
}

func (a *KeyAgreement) isAlgorithmVariant() {}

func (a *KeyAgreement) ToWireInterface() interface{} {
	return a.variant.ToWireInterface()
}

type keyAgreementVariant interface {
	toWire
	isKeyAgreementVariant()
}

func (a *KeyAgreement) GetRaw() *KeyAgreementRaw {
	if alg, ok := a.variant.(*KeyAgreementRaw); ok {
		return alg
	}
	return nil
}
func (a *KeyAgreement) GetWithKeyDerivation() *KeyAgreementWithKeyDerivation {
	if alg, ok := a.variant.(*KeyAgreementWithKeyDerivation); ok {
		return alg
	}
	return nil
}

type KeyAgreementRaw struct {
	RawAlg KeyAgreementRawType
}

func (a *KeyAgreementRaw) isKeyAgreementVariant() {}
func (a *KeyAgreementRaw) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_KeyAgreement_{
			KeyAgreement: &psaalgorithm.Algorithm_KeyAgreement{
				Variant: &psaalgorithm.Algorithm_KeyAgreement_Raw_{
					Raw: psaalgorithm.Algorithm_KeyAgreement_Raw(a.RawAlg),
				},
			},
		},
	}
}

type KeyAgreementWithKeyDerivation struct {
	KaAlg         KeyAgreementRawType
	DerivationAlg *KeyDerivation
}

func (a *KeyAgreementWithKeyDerivation) isKeyAgreementVariant() {}
func (a *KeyAgreementWithKeyDerivation) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_KeyAgreement_{
			KeyAgreement: &psaalgorithm.Algorithm_KeyAgreement{
				Variant: &psaalgorithm.Algorithm_KeyAgreement_WithKeyDerivation_{
					WithKeyDerivation: &psaalgorithm.Algorithm_KeyAgreement_WithKeyDerivation{
						KaAlg:  psaalgorithm.Algorithm_KeyAgreement_Raw(a.KaAlg),
						KdfAlg: a.DerivationAlg.variant.toWireInterfaceSpecific(),
					},
				},
			},
		},
	}
}
