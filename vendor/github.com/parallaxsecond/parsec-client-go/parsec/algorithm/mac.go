// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package algorithm

import (
	"fmt"
	"reflect"

	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaalgorithm"
)

type MACFactory interface {
	HMAC(hashAlg HashAlgorithmType) *Algorithm
	HMACTruncated(hashAlg HashAlgorithmType, macLength uint32) *Algorithm
	CBCMAC() *Algorithm
	CBCMACTruncated(macLength uint32) *Algorithm
	CMAC() *Algorithm
	CMACTruncated(macLength uint32) *Algorithm
}

type macFactory struct{}

func NewMAC() MACFactory {
	return &macFactory{}
}

func (f *macFactory) HMAC(hashAlg HashAlgorithmType) *Algorithm {
	return &Algorithm{
		variant: &MacAlgorithm{
			variant: &MacFullLength{
				variant: &MacFullLengthHmac{
					HashAlg: hashAlg,
				},
			},
		},
	}
}

func (f *macFactory) HMACTruncated(hashAlg HashAlgorithmType, macLength uint32) *Algorithm {
	return &Algorithm{
		variant: &MacAlgorithm{
			variant: &MacTruncated{
				MacAlg: &MacFullLength{
					variant: &MacFullLengthHmac{
						HashAlg: hashAlg,
					},
				},
				MacLength: macLength,
			},
		},
	}
}

func (f *macFactory) CBCMAC() *Algorithm {
	return &Algorithm{
		variant: &MacAlgorithm{
			variant: &MacFullLength{
				variant: &MacFullLengthCbcMac{},
			},
		},
	}
}

func (f *macFactory) CBCMACTruncated(macLength uint32) *Algorithm {
	return &Algorithm{
		variant: &MacAlgorithm{
			variant: &MacTruncated{
				MacAlg: &MacFullLength{
					variant: &MacFullLengthCbcMac{},
				},
				MacLength: macLength,
			},
		},
	}
}

func (f *macFactory) CMAC() *Algorithm {
	return &Algorithm{
		variant: &MacAlgorithm{
			variant: &MacFullLength{
				variant: &MacFullLengthCmac{},
			},
		},
	}
}

func (f *macFactory) CMACTruncated(macLength uint32) *Algorithm {
	return &Algorithm{
		variant: &MacAlgorithm{
			variant: &MacTruncated{
				MacAlg: &MacFullLength{
					variant: &MacFullLengthCbcMac{},
				},
				MacLength: macLength,
			},
		},
	}
}

type MacAlgorithm struct {
	// Assignable from
	// *MacFullLength
	// *MacTruncated
	variant macAlgorithmVariant
}

func (a *MacAlgorithm) isAlgorithmVariant() {}

func (a *MacAlgorithm) ToWireInterface() interface{} {
	return a.variant.ToWireInterface()
}

type macAlgorithmVariant interface {
	toWire
	isMacAlgorithmVariant()
}

type MacFullLength struct {
	// *MacFullLengthHmac
	// *MacFullLengthCbcMac
	// *MacFullLengthCmac

	variant macFullLengthVariant
}

type macFullLengthVariant interface {
	toWire
	toWireInterfaceSpecific() *psaalgorithm.Algorithm_Mac_FullLength
	isMacFullLengthVariant()
}

func (a *MacFullLength) isMacAlgorithmVariant() {}

func (a *MacFullLength) ToWireInterface() interface{} {
	return a.variant.ToWireInterface()
}

type MacTruncated struct {
	MacAlg    *MacFullLength
	MacLength uint32
}

func (a *MacTruncated) isMacAlgorithmVariant() {}

func (a *MacTruncated) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_Mac_{
			Mac: &psaalgorithm.Algorithm_Mac{
				Variant: &psaalgorithm.Algorithm_Mac_Truncated_{
					Truncated: &psaalgorithm.Algorithm_Mac_Truncated{
						MacLength: a.MacLength,
						MacAlg:    a.MacAlg.variant.toWireInterfaceSpecific(),
					},
				},
			},
		},
	}
}

type MacFullLengthHmac struct {
	HashAlg HashAlgorithmType
}

func (a *MacFullLengthHmac) isMacFullLengthVariant() {}
func (a *MacFullLengthHmac) toWireInterfaceSpecific() *psaalgorithm.Algorithm_Mac_FullLength {
	return &psaalgorithm.Algorithm_Mac_FullLength{
		Variant: &psaalgorithm.Algorithm_Mac_FullLength_Hmac_{
			Hmac: &psaalgorithm.Algorithm_Mac_FullLength_Hmac{
				HashAlg: psaalgorithm.Algorithm_Hash(a.HashAlg),
			},
		},
	}
}
func (a *MacFullLengthHmac) ToWireInterface() interface{} {
	return psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_Mac_{
			Mac: &psaalgorithm.Algorithm_Mac{
				Variant: &psaalgorithm.Algorithm_Mac_FullLength_{
					FullLength: a.toWireInterfaceSpecific(),
				},
			},
		},
	}
}

type MacFullLengthCbcMac struct{}

func (a *MacFullLengthCbcMac) isMacFullLengthVariant() {}
func (a *MacFullLengthCbcMac) toWireInterfaceSpecific() *psaalgorithm.Algorithm_Mac_FullLength {
	return &psaalgorithm.Algorithm_Mac_FullLength{
		Variant: &psaalgorithm.Algorithm_Mac_FullLength_CbcMac_{
			CbcMac: &psaalgorithm.Algorithm_Mac_FullLength_CbcMac{},
		},
	}
}
func (a *MacFullLengthCbcMac) ToWireInterface() interface{} {
	return psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_Mac_{
			Mac: &psaalgorithm.Algorithm_Mac{
				Variant: &psaalgorithm.Algorithm_Mac_FullLength_{
					FullLength: a.toWireInterfaceSpecific(),
				},
			},
		},
	}
}

type MacFullLengthCmac struct{}

func (a *MacFullLengthCmac) isMacFullLengthVariant() {}
func (a *MacFullLengthCmac) toWireInterfaceSpecific() *psaalgorithm.Algorithm_Mac_FullLength {
	return &psaalgorithm.Algorithm_Mac_FullLength{
		Variant: &psaalgorithm.Algorithm_Mac_FullLength_Cmac_{
			Cmac: &psaalgorithm.Algorithm_Mac_FullLength_Cmac{},
		},
	}
}
func (a *MacFullLengthCmac) ToWireInterface() interface{} {
	return psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_Mac_{
			Mac: &psaalgorithm.Algorithm_Mac{
				Variant: &psaalgorithm.Algorithm_Mac_FullLength_{
					FullLength: a.toWireInterfaceSpecific(),
				},
			},
		},
	}
}

func newMacFromWire(a *psaalgorithm.Algorithm_Mac) (*MacAlgorithm, error) {
	switch alg := a.Variant.(type) {
	case *psaalgorithm.Algorithm_Mac_FullLength_:
		macfl, err := newMacFullLengthFromWire(alg)
		if err != nil {
			return nil, err
		}
		return &MacAlgorithm{
			variant: macfl,
		}, nil
	case *psaalgorithm.Algorithm_Mac_Truncated_:
		macfl, err := newMacTruncatedFromWire(alg)
		if err != nil {
			return nil, err
		}
		return &MacAlgorithm{
			variant: macfl,
		}, nil

	default:
		return nil, fmt.Errorf("unexpected mac type %v", reflect.TypeOf(alg))
	}
}

func newMacFullLengthFromWire(a *psaalgorithm.Algorithm_Mac_FullLength_) (*MacFullLength, error) {
	var macalg macFullLengthVariant
	switch alg := a.FullLength.Variant.(type) {
	case *psaalgorithm.Algorithm_Mac_FullLength_CbcMac_:
		macalg = &MacFullLengthCbcMac{}
	case *psaalgorithm.Algorithm_Mac_FullLength_Cmac_:
		macalg = &MacFullLengthCmac{}
	case *psaalgorithm.Algorithm_Mac_FullLength_Hmac_:
		macalg = &MacFullLengthHmac{
			HashAlg: HashAlgorithmType(alg.Hmac.HashAlg),
		}
	default:
		return nil, fmt.Errorf("unexpected full length mac type %v", reflect.TypeOf(alg))
	}
	return &MacFullLength{
		variant: macalg,
	}, nil
}

func newMacTruncatedFromWire(a *psaalgorithm.Algorithm_Mac_Truncated_) (*MacTruncated, error) {
	macalg, err := newMacFullLengthFromWire(&psaalgorithm.Algorithm_Mac_FullLength_{FullLength: a.Truncated.MacAlg})
	if err != nil {
		return nil, err
	}
	return &MacTruncated{
		MacAlg:    macalg,
		MacLength: a.Truncated.MacLength,
	}, err
}
