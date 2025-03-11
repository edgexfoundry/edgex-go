// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package algorithm

import (
	"fmt"
	"reflect"

	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaalgorithm"
)

type AeadFactory interface {
	Aead(algType AeadAlgorithmType) *Algorithm
	AeadShortenedTag(algType AeadAlgorithmType, tagLength uint32) *Algorithm
}

type aeadFactory struct{}

func NewAead() AeadFactory {
	return &aeadFactory{}
}
func (a *aeadFactory) Aead(algType AeadAlgorithmType) *Algorithm {
	return &Algorithm{
		variant: &AeadAlgorithm{
			variant: &AeadAlgorithmDefaultLengthTag{
				AeadAlg: algType,
			},
		},
	}
}

func (a *aeadFactory) AeadShortenedTag(algType AeadAlgorithmType, tagLength uint32) *Algorithm {
	return &Algorithm{
		variant: &AeadAlgorithm{
			variant: &AeadAlgorithmShortenedTag{
				AeadAlg: algType,
			},
		},
	}
}

type AeadAlgorithmType uint32

const (
	AeadAlgorithmNODEFAULTTAG     AeadAlgorithmType = 0
	AeadAlgorithmCCM              AeadAlgorithmType = 1
	AeadAlgorithmGCM              AeadAlgorithmType = 2
	AeadAlgorithmChacha20Poly1305 AeadAlgorithmType = 4
)

type AeadAlgorithm struct {
	//	*AeadWithDefaultLengthTag
	//	*AeadWithShortenedTag
	variant aeadAlgorithmVariant
}

type aeadAlgorithmVariant interface {
	toWire
	isAeadAlgorithmVariant()
}

func (a AeadAlgorithm) isAlgorithmVariant() {}

func (a *AeadAlgorithm) ToWireInterface() interface{} {
	return a.variant.ToWireInterface()
}
func (a *AeadAlgorithm) GetAeadDefaultLengthTag() *AeadAlgorithmDefaultLengthTag {
	if alg, ok := a.variant.(*AeadAlgorithmDefaultLengthTag); ok {
		return alg
	}
	return nil
}
func (a *AeadAlgorithm) GetAeadShortenedTag() *AeadAlgorithmShortenedTag {
	if alg, ok := a.variant.(*AeadAlgorithmShortenedTag); ok {
		return alg
	}
	return nil
}

type AeadAlgorithmDefaultLengthTag struct {
	AeadAlg AeadAlgorithmType
}

func (a *AeadAlgorithmDefaultLengthTag) isAeadAlgorithmVariant() {}

func (a *AeadAlgorithmDefaultLengthTag) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_Aead_{
			Aead: &psaalgorithm.Algorithm_Aead{
				Variant: &psaalgorithm.Algorithm_Aead_AeadWithDefaultLengthTag_{
					AeadWithDefaultLengthTag: psaalgorithm.Algorithm_Aead_AeadWithDefaultLengthTag(a.AeadAlg),
				},
			},
		},
	}
}

type AeadAlgorithmShortenedTag struct {
	AeadAlg   AeadAlgorithmType
	TagLength uint32
}

func (a *AeadAlgorithmShortenedTag) isAeadAlgorithmVariant() {}

func (a *AeadAlgorithmShortenedTag) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_Aead_{
			Aead: &psaalgorithm.Algorithm_Aead{
				Variant: &psaalgorithm.Algorithm_Aead_AeadWithShortenedTag_{
					AeadWithShortenedTag: &psaalgorithm.Algorithm_Aead_AeadWithShortenedTag{
						AeadAlg:   psaalgorithm.Algorithm_Aead_AeadWithDefaultLengthTag(a.AeadAlg),
						TagLength: a.TagLength,
					},
				},
			},
		},
	}
}

func newAeadFromWire(a *psaalgorithm.Algorithm_Aead) (*AeadAlgorithm, error) {
	switch linealg := a.Variant.(type) {
	case *psaalgorithm.Algorithm_Aead_AeadWithDefaultLengthTag_:
		return &AeadAlgorithm{
			variant: &AeadAlgorithmDefaultLengthTag{
				AeadAlg: AeadAlgorithmType(linealg.AeadWithDefaultLengthTag),
			},
		}, nil
	case *psaalgorithm.Algorithm_Aead_AeadWithShortenedTag_:
		return &AeadAlgorithm{
			variant: &AeadAlgorithmShortenedTag{
				AeadAlg:   AeadAlgorithmType(linealg.AeadWithShortenedTag.AeadAlg),
				TagLength: linealg.AeadWithShortenedTag.TagLength,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unexpected type encountered decoding aead algorithm: %v", reflect.TypeOf(linealg))
	}
}
