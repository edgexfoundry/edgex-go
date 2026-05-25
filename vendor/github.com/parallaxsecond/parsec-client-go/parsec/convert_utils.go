// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package parsec

import (
	"fmt"
	"reflect"

	"github.com/parallaxsecond/parsec-client-go/interface/auth"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/listauthenticators"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaalgorithm"
	"github.com/parallaxsecond/parsec-client-go/parsec/algorithm"
)

func newAuthenticatorInfoFromOp(inf *listauthenticators.AuthenticatorInfo) (*AuthenticatorInfo, error) {
	authid, err := auth.NewAuthenticationTypeFromU32(inf.Id)
	if err != nil {
		return nil, err
	}
	return &AuthenticatorInfo{
		ID:          AuthenticatorType(authid),
		Description: inf.Description,
		VersionMaj:  inf.VersionMaj,
		VersionMin:  inf.VersionMin,
		VersionRev:  inf.VersionRev,
	}, nil
}

func hashAlgToWire(h algorithm.HashAlgorithmType) psaalgorithm.Algorithm_Hash {
	return psaalgorithm.Algorithm_Hash(h)
}

func algAsymmetricSigToWire(a *algorithm.AsymmetricSignatureAlgorithm) (*psaalgorithm.Algorithm_AsymmetricSignature, error) {
	aif := a.ToWireInterface()
	alg, ok := a.ToWireInterface().(*psaalgorithm.Algorithm)
	if !ok {
		return nil, fmt.Errorf("unexpected type expecting *psaalgorithm.Algorithm, got %v", reflect.TypeOf(aif))
	}
	varalg := alg.GetAsymmetricSignature()
	if varalg == nil {
		return nil, fmt.Errorf("expected *psaalgorithm.Algorithm_AsymmetricSignature, but got nil")
	}
	return varalg, nil
}

func algAsymmetricEncryptionAlgToWire(a *algorithm.AsymmetricEncryptionAlgorithm) (*psaalgorithm.Algorithm_AsymmetricEncryption, error) {
	alg, ok := a.ToWireInterface().(*psaalgorithm.Algorithm)
	if !ok {
		return nil, fmt.Errorf("unexpected type expecting *psaalgorithm.Algorithm, got %v", reflect.TypeOf(alg))
	}
	varalg := alg.GetAsymmetricEncryption()
	if varalg == nil {
		return nil, fmt.Errorf("expected *psaalgorithm.Algorithm_AsymmetricEncryption, but got nil")
	}
	return varalg, nil
}

func algCipherAlgToWire(a *algorithm.Cipher) (psaalgorithm.Algorithm_Cipher, error) {
	alg, ok := a.ToWireInterface().(*psaalgorithm.Algorithm)
	if !ok {
		return psaalgorithm.Algorithm_CIPHER_NONE, fmt.Errorf("unexpected type expecting *psaalgorithm.Algorithm, got %v", reflect.TypeOf(alg))
	}
	varalg := alg.GetCipher()
	if varalg == psaalgorithm.Algorithm_CIPHER_NONE {
		return psaalgorithm.Algorithm_CIPHER_NONE, fmt.Errorf("expected *psaalgorithm.Algorithm_AsymmetricEncryption, but got nil")
	}
	return varalg, nil
}

func algAeadAlgToWire(a *algorithm.AeadAlgorithm) (*psaalgorithm.Algorithm_Aead, error) {
	alg, ok := a.ToWireInterface().(*psaalgorithm.Algorithm)
	if !ok {
		return nil, fmt.Errorf("unexpected type expecting *psaalgorithm.Algorithm, got %v", reflect.TypeOf(alg))
	}
	varalg := alg.GetAead()
	if varalg == nil {
		return nil, fmt.Errorf("expected *psaalgorithm.Algorithm_Aead, but got nil")
	}
	return varalg, nil
}

func algMacAlgToWire(a *algorithm.MacAlgorithm) (*psaalgorithm.Algorithm_Mac, error) {
	alg, ok := a.ToWireInterface().(*psaalgorithm.Algorithm)
	if !ok {
		return nil, fmt.Errorf("unexpected type expecting *psaalgorithm.Algorithm, got %v", reflect.TypeOf(alg))
	}
	varalg := alg.GetMac()
	if varalg == nil {
		return nil, fmt.Errorf("expected *psaalgorithm.Algorithm_Mac, but got nil")
	}
	return varalg, nil
}

func algKeyAgreementRawAlgToWire(a *algorithm.KeyAgreementRaw) (*psaalgorithm.Algorithm_KeyAgreement, error) {
	alg, ok := a.ToWireInterface().(*psaalgorithm.Algorithm)
	if !ok {
		return nil, fmt.Errorf("unexpected type expecting *psaalgorithm.Algorithm, got %v", reflect.TypeOf(alg))
	}
	varalg := alg.GetKeyAgreement()
	if varalg == nil {
		return nil, fmt.Errorf("expected *psaalgorithm.Algorithm_KeyAgreement_Raw, but got nil")
	}
	return varalg, nil
}
