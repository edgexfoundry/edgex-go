// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package parsec

import (
	"fmt"
	"reflect"

	"github.com/parallaxsecond/parsec-client-go/interface/operations/listkeys"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaalgorithm"
	"github.com/parallaxsecond/parsec-client-go/interface/operations/psakeyattributes"
	"github.com/parallaxsecond/parsec-client-go/parsec/algorithm"
)

type ToWire interface {
	ToWireInterface() interface{}
}

type UsageFlags struct {
	Export        bool
	Copy          bool
	Cache         bool
	Encrypt       bool
	Decrypt       bool
	SignMessage   bool
	VerifyMessage bool
	SignHash      bool
	VerifyHash    bool
	Derive        bool
}

func (u *UsageFlags) toNativeWireInterface() *psakeyattributes.UsageFlags {
	return &psakeyattributes.UsageFlags{
		Export:        u.Export,
		Copy:          u.Copy,
		Cache:         u.Cache,
		Encrypt:       u.Encrypt,
		Decrypt:       u.Decrypt,
		SignMessage:   u.SignMessage,
		VerifyMessage: u.VerifyMessage,
		VerifyHash:    u.VerifyHash,
		SignHash:      u.SignHash,
		Derive:        u.Derive,
	}
}

type KeyPolicy struct {
	KeyUsageFlags *UsageFlags
	KeyAlgorithm  *algorithm.Algorithm
}

func (kp *KeyPolicy) toNativeWireInterface() *psakeyattributes.KeyPolicy {
	kaif := kp.KeyAlgorithm.ToWireInterface()
	if kaif == nil {
		fmt.Println("no wire alg from kp.KeyAlgorithm")
		return nil
	}
	ka, ok := kaif.(*psaalgorithm.Algorithm)
	if !ok {
		fmt.Printf("kaif is wrong type, got %v\n", reflect.TypeOf(kaif))
		return nil
	}

	return &psakeyattributes.KeyPolicy{
		KeyAlgorithm:  ka,
		KeyUsageFlags: kp.KeyUsageFlags.toNativeWireInterface(),
	}
}

type KeyAttributes struct {
	KeyType   *KeyType
	KeyBits   uint32
	KeyPolicy *KeyPolicy
}

func newKeyAttributesFromOp(ka *psakeyattributes.KeyAttributes) (*KeyAttributes, error) { //nolint:unparam
	return &KeyAttributes{
		KeyBits: ka.KeyBits,
		// KeyPolicy: ,
	}, nil
	// TODO finish this
}

func (ka *KeyAttributes) toWireInterface() (*psakeyattributes.KeyAttributes, error) {
	keytypeif := ka.KeyType.ToWireInterface()
	if keytypeif == nil {
		return nil, fmt.Errorf("nil keytype returned for wire interface")
	}
	keytype, ok := keytypeif.(*psakeyattributes.KeyType)
	if !ok {
		return nil, fmt.Errorf("incorrect type returned for keytype")
	}
	return &psakeyattributes.KeyAttributes{
		KeyBits:   ka.KeyBits,
		KeyType:   keytype,
		KeyPolicy: ka.KeyPolicy.toNativeWireInterface(),
	}, nil
}

type KeyInfo struct {
	ProviderID ProviderID
	Name       string
	Attributes *KeyAttributes
}

func newKeyInfoFromOp(wireinf *listkeys.KeyInfo) (*KeyInfo, error) {
	ka, err := newKeyAttributesFromOp(wireinf.Attributes)
	if err != nil {
		return nil, err
	}
	return &KeyInfo{
		ProviderID: ProviderID(wireinf.ProviderId),
		Name:       wireinf.Name,
		Attributes: ka,
	}, nil
}
