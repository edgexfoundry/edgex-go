// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package algorithm

import (
	"fmt"
	"reflect"

	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaalgorithm"
)

type AsymmetricSignatureFactory interface {
	RsaPkcs1V15Sign(hashAlg HashAlgorithmType) *Algorithm
	RsaPkcs1V15SignAny() *Algorithm
	RsaPkcs1V15SignRaw() *Algorithm
	RsaPss(hashAlg HashAlgorithmType) *Algorithm
	RsaPssAny() *Algorithm
	Ecdsa(hashAlg HashAlgorithmType) *Algorithm
	EcdsaAny() *Algorithm
	DeterministicEcdsa(hashAlg HashAlgorithmType) *Algorithm
	DeterministicEcdsaAny() *Algorithm
}

type asymmetricSignatureFactory struct{}

func NewAsymmetricSignature() AsymmetricSignatureFactory {
	return &asymmetricSignatureFactory{}
}

func (a *asymmetricSignatureFactory) RsaPkcs1V15Sign(hashAlg HashAlgorithmType) *Algorithm {
	return &Algorithm{
		variant: &AsymmetricSignatureAlgorithm{
			variant: &AsymmetricSignatureRsaPkcs1V15Sign{
				SignHash: &AsymmetricSignatureSignHash{
					variant: &AsymmetricSignatureSignHashSpecific{
						HashAlg: hashAlg,
					},
				},
			},
		},
	}
}
func (a *asymmetricSignatureFactory) RsaPkcs1V15SignAny() *Algorithm {
	return &Algorithm{
		variant: &AsymmetricSignatureAlgorithm{
			variant: &AsymmetricSignatureRsaPkcs1V15Sign{
				SignHash: &AsymmetricSignatureSignHash{
					variant: &AsymmetricSignatureSignHashAny{},
				},
			},
		},
	}
}

func (a *asymmetricSignatureFactory) RsaPkcs1V15SignRaw() *Algorithm {
	return &Algorithm{
		variant: &AsymmetricSignatureAlgorithm{
			variant: &AsymmetricSignatureRsaPkcs1V15SignRaw{},
		},
	}
}

func (a *asymmetricSignatureFactory) RsaPss(hashAlg HashAlgorithmType) *Algorithm {
	return &Algorithm{
		variant: &AsymmetricSignatureAlgorithm{
			variant: &AsymmetricSignatureRsaPss{
				SignHash: &AsymmetricSignatureSignHash{
					variant: &AsymmetricSignatureSignHashSpecific{
						HashAlg: hashAlg,
					},
				},
			},
		},
	}
}

func (a *asymmetricSignatureFactory) RsaPssAny() *Algorithm {
	return &Algorithm{
		variant: &AsymmetricSignatureAlgorithm{
			variant: &AsymmetricSignatureRsaPss{
				SignHash: &AsymmetricSignatureSignHash{
					variant: &AsymmetricSignatureSignHashAny{},
				},
			},
		},
	}
}

func (a *asymmetricSignatureFactory) Ecdsa(hashAlg HashAlgorithmType) *Algorithm {
	return &Algorithm{
		variant: &AsymmetricSignatureAlgorithm{
			variant: &AsymmetricSignatureEcdsa{
				SignHash: &AsymmetricSignatureSignHash{
					variant: &AsymmetricSignatureSignHashSpecific{
						HashAlg: hashAlg,
					},
				},
			},
		},
	}
}

func (a *asymmetricSignatureFactory) EcdsaAny() *Algorithm {
	return &Algorithm{
		variant: &AsymmetricSignatureAlgorithm{
			variant: &AsymmetricSignatureEcdsaAny{},
		},
	}
}

func (a *asymmetricSignatureFactory) DeterministicEcdsa(hashAlg HashAlgorithmType) *Algorithm {
	return &Algorithm{
		variant: &AsymmetricSignatureAlgorithm{
			variant: &AsymmetricSignatureDeterministicEcdsa{
				SignHash: &AsymmetricSignatureSignHash{
					variant: &AsymmetricSignatureSignHashSpecific{
						HashAlg: hashAlg,
					},
				},
			},
		},
	}
}

func (a *asymmetricSignatureFactory) DeterministicEcdsaAny() *Algorithm {
	return &Algorithm{
		variant: &AsymmetricSignatureAlgorithm{
			variant: &AsymmetricSignatureDeterministicEcdsa{
				SignHash: &AsymmetricSignatureSignHash{
					variant: &AsymmetricSignatureSignHashAny{},
				},
			},
		},
	}
}

// *AsymmetricSignatureRsaPkcs1V15Sign
// *AsymmetricSignatureRsaPkcs1V15SignRaw
// *AsymmetricSignatureRsaPss
// *AsymmetricSignatureEcdsa
// *AsymmetricSignatureEcdsaAny
// *AsymmetricSignatureDeterministicEcdsa
type AsymmetricSignatureAlgorithm struct {
	variant asymmetricAlgorithmVariant
}

type asymmetricAlgorithmVariant interface {
	isAsymmetricAlgorithmVariant()
	ToWireInterface() interface{}
}

func (a AsymmetricSignatureAlgorithm) isAlgorithmVariant() {}

func (a *AsymmetricSignatureAlgorithm) ToWireInterface() interface{} {
	if a.variant == nil {
		return nil
	}
	return a.variant.ToWireInterface()
}

func (a *AsymmetricSignatureAlgorithm) GetRsaPkcs1V15Sign() *AsymmetricSignatureRsaPkcs1V15Sign {
	if alg, ok := a.variant.(*AsymmetricSignatureRsaPkcs1V15Sign); ok {
		return alg
	}
	return nil
}
func (a *AsymmetricSignatureAlgorithm) GetRsaPkcs1V15SignRaw() *AsymmetricSignatureRsaPkcs1V15SignRaw {
	if alg, ok := a.variant.(*AsymmetricSignatureRsaPkcs1V15SignRaw); ok {
		return alg
	}
	return nil
}
func (a *AsymmetricSignatureAlgorithm) GetRsaPss() *AsymmetricSignatureRsaPss {
	if alg, ok := a.variant.(*AsymmetricSignatureRsaPss); ok {
		return alg
	}
	return nil
}
func (a *AsymmetricSignatureAlgorithm) GetEcdsa() *AsymmetricSignatureEcdsa {
	if alg, ok := a.variant.(*AsymmetricSignatureEcdsa); ok {
		return alg
	}
	return nil
}
func (a *AsymmetricSignatureAlgorithm) GetEcdsaAny() *AsymmetricSignatureEcdsaAny {
	if alg, ok := a.variant.(*AsymmetricSignatureEcdsaAny); ok {
		return alg
	}
	return nil
}
func (a *AsymmetricSignatureAlgorithm) GetDeterministicEcdsa() *AsymmetricSignatureDeterministicEcdsa {
	if alg, ok := a.variant.(*AsymmetricSignatureDeterministicEcdsa); ok {
		return alg
	}
	return nil
}

type AsymmetricSignatureRsaPkcs1V15Sign struct {
	SignHash *AsymmetricSignatureSignHash
}

func (a AsymmetricSignatureRsaPkcs1V15Sign) isAsymmetricAlgorithmVariant() {}
func (a *AsymmetricSignatureRsaPkcs1V15Sign) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_AsymmetricSignature_{
			AsymmetricSignature: &psaalgorithm.Algorithm_AsymmetricSignature{
				Variant: &psaalgorithm.Algorithm_AsymmetricSignature_RsaPkcs1V15Sign_{
					RsaPkcs1V15Sign: &psaalgorithm.Algorithm_AsymmetricSignature_RsaPkcs1V15Sign{
						HashAlg: a.SignHash.toWireInterfaceSpecific(),
					},
				},
			},
		},
	}
}

type AsymmetricSignatureRsaPkcs1V15SignRaw struct {
}

func (a *AsymmetricSignatureRsaPkcs1V15SignRaw) isAsymmetricAlgorithmVariant() {}

func (a *AsymmetricSignatureRsaPkcs1V15SignRaw) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_AsymmetricSignature_{
			AsymmetricSignature: &psaalgorithm.Algorithm_AsymmetricSignature{
				Variant: &psaalgorithm.Algorithm_AsymmetricSignature_RsaPkcs1V15SignRaw_{
					RsaPkcs1V15SignRaw: &psaalgorithm.Algorithm_AsymmetricSignature_RsaPkcs1V15SignRaw{},
				},
			},
		},
	}
}

type AsymmetricSignatureRsaPss struct {
	SignHash *AsymmetricSignatureSignHash
}

func (a *AsymmetricSignatureRsaPss) isAsymmetricAlgorithmVariant() {}

func (a *AsymmetricSignatureRsaPss) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_AsymmetricSignature_{
			AsymmetricSignature: &psaalgorithm.Algorithm_AsymmetricSignature{
				Variant: &psaalgorithm.Algorithm_AsymmetricSignature_RsaPss_{
					RsaPss: &psaalgorithm.Algorithm_AsymmetricSignature_RsaPss{
						HashAlg: a.SignHash.variant.toWireInterfaceSpecific(),
					},
				},
			},
		},
	}
}

type AsymmetricSignatureEcdsa struct {
	SignHash *AsymmetricSignatureSignHash
}

func (a *AsymmetricSignatureEcdsa) isAsymmetricAlgorithmVariant() {}

func (a *AsymmetricSignatureEcdsa) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_AsymmetricSignature_{
			AsymmetricSignature: &psaalgorithm.Algorithm_AsymmetricSignature{
				Variant: &psaalgorithm.Algorithm_AsymmetricSignature_Ecdsa_{
					Ecdsa: &psaalgorithm.Algorithm_AsymmetricSignature_Ecdsa{
						HashAlg: a.SignHash.variant.toWireInterfaceSpecific(),
					},
				},
			},
		},
	}
}

type AsymmetricSignatureEcdsaAny struct {
}

func (a *AsymmetricSignatureEcdsaAny) isAsymmetricAlgorithmVariant() {}

func (a *AsymmetricSignatureEcdsaAny) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_AsymmetricSignature_{
			AsymmetricSignature: &psaalgorithm.Algorithm_AsymmetricSignature{
				Variant: &psaalgorithm.Algorithm_AsymmetricSignature_EcdsaAny_{
					EcdsaAny: &psaalgorithm.Algorithm_AsymmetricSignature_EcdsaAny{},
				},
			},
		},
	}
}

type AsymmetricSignatureDeterministicEcdsa struct {
	SignHash *AsymmetricSignatureSignHash
}

func (a *AsymmetricSignatureDeterministicEcdsa) isAsymmetricAlgorithmVariant() {}

func (a *AsymmetricSignatureDeterministicEcdsa) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_AsymmetricSignature_{
			AsymmetricSignature: &psaalgorithm.Algorithm_AsymmetricSignature{
				Variant: &psaalgorithm.Algorithm_AsymmetricSignature_DeterministicEcdsa_{
					DeterministicEcdsa: &psaalgorithm.Algorithm_AsymmetricSignature_DeterministicEcdsa{
						HashAlg: a.SignHash.variant.toWireInterfaceSpecific(),
					},
				},
			},
		},
	}
}

func newAsymmetricSignatureFromWire(a *psaalgorithm.Algorithm_AsymmetricSignature) (*AsymmetricSignatureAlgorithm, error) {
	var sig asymmetricAlgorithmVariant
	switch linealg := a.Variant.(type) {
	case *psaalgorithm.Algorithm_AsymmetricSignature_RsaPkcs1V15Sign_:
		shash, err := newAsymmetricSignatureSignHashFromWire(linealg.RsaPkcs1V15Sign.HashAlg)
		if err != nil {
			return nil, err
		}
		sig = &AsymmetricSignatureRsaPkcs1V15Sign{
			SignHash: shash,
		}
	case *psaalgorithm.Algorithm_AsymmetricSignature_RsaPkcs1V15SignRaw_:
		sig = &AsymmetricSignatureRsaPkcs1V15SignRaw{}
	case *psaalgorithm.Algorithm_AsymmetricSignature_RsaPss_:
		shash, err := newAsymmetricSignatureSignHashFromWire(linealg.RsaPss.HashAlg)
		if err != nil {
			return nil, err
		}
		sig = &AsymmetricSignatureRsaPss{
			SignHash: shash,
		}
	case *psaalgorithm.Algorithm_AsymmetricSignature_Ecdsa_:
		shash, err := newAsymmetricSignatureSignHashFromWire(linealg.Ecdsa.HashAlg)
		if err != nil {
			return nil, err
		}
		sig = &AsymmetricSignatureEcdsa{
			SignHash: shash,
		}
	case *psaalgorithm.Algorithm_AsymmetricSignature_EcdsaAny_:
		sig = &AsymmetricSignatureEcdsaAny{}
	case *psaalgorithm.Algorithm_AsymmetricSignature_DeterministicEcdsa_:
		shash, err := newAsymmetricSignatureSignHashFromWire(linealg.DeterministicEcdsa.HashAlg)
		if err != nil {
			return nil, err
		}
		sig = &AsymmetricSignatureDeterministicEcdsa{
			SignHash: shash,
		}
	default:
		return nil, fmt.Errorf("unexpected AsymmetricSignatureAlgorithm variant %v", reflect.TypeOf(linealg))
	}

	return &AsymmetricSignatureAlgorithm{
		variant: sig,
	}, nil
}

type AsymmetricSignatureSignHash struct {
	// Can be assigned to by
	// AsymmetricSignatureSignHashVariantSpecific and
	// AsymmetricSignatureSignHashVariantAny
	variant asymmetricSignatureSignHashVariant
}

func (x *AsymmetricSignatureSignHash) toWireInterfaceSpecific() *psaalgorithm.Algorithm_AsymmetricSignature_SignHash {
	return x.variant.toWireInterfaceSpecific()
}

func newAsymmetricSignatureSignHashFromWire(a *psaalgorithm.Algorithm_AsymmetricSignature_SignHash) (*AsymmetricSignatureSignHash, error) {
	// Types that are assignable to Variant:
	//	*Algorithm_AsymmetricSignature_SignHash_Any_
	//	*Algorithm_AsymmetricSignature_SignHash_Specific
	switch linealg := a.Variant.(type) {
	case *psaalgorithm.Algorithm_AsymmetricSignature_SignHash_Any_:
		return &AsymmetricSignatureSignHash{
			variant: &AsymmetricSignatureSignHashAny{},
		}, nil
	case *psaalgorithm.Algorithm_AsymmetricSignature_SignHash_Specific:
		return &AsymmetricSignatureSignHash{
			variant: &AsymmetricSignatureSignHashSpecific{
				HashAlg: HashAlgorithmType(linealg.Specific),
			},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected AsymmetricSignatureSignHash variant type %v", reflect.TypeOf(linealg))
	}
}

type asymmetricSignatureSignHashVariant interface {
	isAsymmetricSignatureSignHashVariant()
	toWireInterfaceSpecific() *psaalgorithm.Algorithm_AsymmetricSignature_SignHash
}

func (x *AsymmetricSignatureSignHash) GetAny() *AsymmetricSignatureSignHashAny {
	if x, ok := x.variant.(*AsymmetricSignatureSignHashAny); ok {
		return x
	}
	return nil
}

func (x *AsymmetricSignatureSignHash) GetSpecific() HashAlgorithmType {
	if x, ok := x.variant.(*AsymmetricSignatureSignHashSpecific); ok {
		return x.HashAlg
	}
	return HashAlgorithmTypeNONE
}

type AsymmetricSignatureSignHashAny struct{}

func (a *AsymmetricSignatureSignHashAny) isAsymmetricSignatureSignHashVariant() {}

func (a *AsymmetricSignatureSignHashAny) toWireInterfaceSpecific() *psaalgorithm.Algorithm_AsymmetricSignature_SignHash {
	return &psaalgorithm.Algorithm_AsymmetricSignature_SignHash{
		Variant: &psaalgorithm.Algorithm_AsymmetricSignature_SignHash_Any_{},
	}
}

type AsymmetricSignatureSignHashSpecific struct {
	HashAlg HashAlgorithmType
}

func (a *AsymmetricSignatureSignHashSpecific) isAsymmetricSignatureSignHashVariant() {}

func (a *AsymmetricSignatureSignHashSpecific) toWireInterfaceSpecific() *psaalgorithm.Algorithm_AsymmetricSignature_SignHash {
	return &psaalgorithm.Algorithm_AsymmetricSignature_SignHash{
		Variant: &psaalgorithm.Algorithm_AsymmetricSignature_SignHash_Specific{
			Specific: psaalgorithm.Algorithm_Hash(a.HashAlg),
		},
	}
}
