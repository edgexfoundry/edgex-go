// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package algorithm

import (
	"fmt"

	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaalgorithm"
)

type HashAlgorithmType uint32

const (
	HashAlgorithmTypeNONE HashAlgorithmType = 0 // This default variant should not be used.
	// Deprecated: Do not use.
	HashAlgorithmTypeMD2 HashAlgorithmType = 1
	// Deprecated: Do not use.
	HashAlgorithmTypeMD4 HashAlgorithmType = 2
	// Deprecated: Do not use.
	HashAlgorithmTypeMD5       HashAlgorithmType = 3
	HashAlgorithmTypeRIPEMD160 HashAlgorithmType = 4
	// Deprecated: Do not use.
	HashAlgorithmTypeSHA1       HashAlgorithmType = 5
	HashAlgorithmTypeSHA224     HashAlgorithmType = 6
	HashAlgorithmTypeSHA256     HashAlgorithmType = 7
	HashAlgorithmTypeSHA384     HashAlgorithmType = 8
	HashAlgorithmTypeSHA512     HashAlgorithmType = 9
	HashAlgorithmTypeSHA512_224 HashAlgorithmType = 10
	HashAlgorithmTypeSHA512_256 HashAlgorithmType = 11
	HashAlgorithmTypeSHA3_224   HashAlgorithmType = 12
	HashAlgorithmTypeSHA3_256   HashAlgorithmType = 13
	HashAlgorithmTypeSHA3_384   HashAlgorithmType = 14
	HashAlgorithmTypeSHA3_512   HashAlgorithmType = 15
)

//nolint:gocyclo
func (a HashAlgorithmType) String() string {
	switch a {
	case HashAlgorithmTypeNONE:
		return "HASH_NONE"
	case HashAlgorithmTypeMD2:
		return "MD2"
	case HashAlgorithmTypeMD4:
		return "MD4"
	case HashAlgorithmTypeMD5:
		return "MD5"
	case HashAlgorithmTypeRIPEMD160:
		return "RIPEMD160"
	case HashAlgorithmTypeSHA1:
		return "SHA_1"
	case HashAlgorithmTypeSHA224:
		return "SHA_224"
	case HashAlgorithmTypeSHA256:
		return "SHA_256"
	case HashAlgorithmTypeSHA384:
		return "SHA_384"
	case HashAlgorithmTypeSHA512:
		return "SHA_512"
	case HashAlgorithmTypeSHA512_224:
		return "SHA_512_224"
	case HashAlgorithmTypeSHA512_256:
		return "SHA_512_256"
	case HashAlgorithmTypeSHA3_224:
		return "SHA3_224"
	case HashAlgorithmTypeSHA3_256:
		return "SHA3_256"
	case HashAlgorithmTypeSHA3_384:
		return "SHA3_384"
	case HashAlgorithmTypeSHA3_512:
		return "SHA3_512"
	}
	return ""
}

type HashAlgorithm struct {
	HashAlg HashAlgorithmType
}

func NewHashAlgorithm(h HashAlgorithmType) *Algorithm {
	return &Algorithm{
		variant: &HashAlgorithm{
			HashAlg: h,
		},
	}
}

func (a HashAlgorithm) String() string {
	return a.HashAlg.String()
}

func (a HashAlgorithm) isAlgorithmVariant() {}
func (a *HashAlgorithm) ToWireInterface() interface{} {
	return psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_Hash_{
			// We've defined HashAlg to be same as protoc interface so we can cast safely
			Hash: psaalgorithm.Algorithm_Hash(a.HashAlg),
		},
	}
}

func newHashFromWire(a psaalgorithm.Algorithm_Hash) (algorithmVariant, error) {
	return nil, fmt.Errorf("not implemented")
}
