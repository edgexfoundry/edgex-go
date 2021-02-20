//
// Copyright (c) 2021 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package secretstore

import (
	"context"
	"crypto/rand"
	"encoding/base64"
)

const randomBytesLength = 33 // 264 bits of entropy

// CredentialGenerator is the interface for pluggable password generators
type CredentialGenerator interface {
	Generate(ctx context.Context) (string, error)
}

type defaultCredentialGenerator struct{}

// NewDefaultCredentialGenerator generates random passwords as base64-encoded strings
func NewDefaultCredentialGenerator() CredentialGenerator {
	return &defaultCredentialGenerator{}
}

// Generate implementation returns base64-encoded randomBytesLength random bytes
func (cg *defaultCredentialGenerator) Generate(_ context.Context) (string, error) {
	randomBytes := make([]byte, randomBytesLength)
	_, err := rand.Read(randomBytes) // all of salt guaranteed to be filled if err==nil
	if err != nil {
		return "", err
	}
	newCredential := base64.StdEncoding.EncodeToString(randomBytes)
	return newCredential, nil
}
