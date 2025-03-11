// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package requests

// ProviderID for providers
type ProviderID uint8

// Provider UUIDs
const (
	ProviderCore           ProviderID = 0
	ProviderMBed           ProviderID = 1
	ProviderPKCS11         ProviderID = 2
	ProviderTPM            ProviderID = 3
	ProviderTrustedService ProviderID = 4
)

// HasCrypto returns true if the provider supports crypto
func (p ProviderID) HasCrypto() bool {
	return p.IsValid() && p != ProviderCore
}

func (p ProviderID) IsValid() bool {
	return p >= ProviderCore && p <= ProviderTrustedService
}

func (p ProviderID) String() string {
	switch p {
	case ProviderCore:
		return "Core"
	case ProviderMBed:
		return "MBed"
	case ProviderPKCS11:
		return "PKCS11"
	case ProviderTPM:
		return "TPM"
	case ProviderTrustedService:
		return "TrustedService"
	default:
		return "Unknown"
	}
}
