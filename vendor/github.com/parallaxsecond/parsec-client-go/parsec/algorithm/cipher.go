// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package algorithm

import (
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaalgorithm"
)

type CipherModeType int32

const (
	CipherModeCIPHERNONE   CipherModeType = 0 // This default variant should not be used.
	CipherModeSTREAMCIPHER CipherModeType = 1
	CipherModeCTR          CipherModeType = 2
	CipherModeCFB          CipherModeType = 3
	CipherModeOFB          CipherModeType = 4
	CipherModeXTS          CipherModeType = 5
	CipherModeECBNOPADDING CipherModeType = 6
	CipherModeCBCNOPADDING CipherModeType = 7
	CipherModeCBCPKCS7     CipherModeType = 8
)

type Cipher struct {
	Mode CipherModeType
}

func (c *Cipher) isAlgorithmVariant() {}

func (c *Cipher) ToWireInterface() interface{} {
	return &psaalgorithm.Algorithm{
		Variant: &psaalgorithm.Algorithm_Cipher_{
			// We've defined cipher mode to be same as protoc interface so we can cast safely
			Cipher: psaalgorithm.Algorithm_Cipher(c.Mode),
		},
	}
}

func NewCipher(mode CipherModeType) *Algorithm {
	return &Algorithm{
		variant: &Cipher{
			Mode: mode,
		},
	}
}

func newCipherFromWire(a psaalgorithm.Algorithm_Cipher) (*Cipher, error) { //nolint:unparam
	return &Cipher{
		Mode: CipherModeType(a),
	}, nil
}
