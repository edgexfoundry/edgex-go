// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package parsec

import "github.com/parallaxsecond/parsec-client-go/interface/operations/psakeyattributes"

type KeyTypeFactory interface {
	RawData() *KeyType
	Hmac() *KeyType
	Derive() *KeyType
	Aes() *KeyType
	Des() *KeyType
	Camellia() *KeyType
	Arc4() *KeyType
	Chacha20() *KeyType
	RsaPublicKey() *KeyType
	RsaKeyPair() *KeyType
	EccKeyPair(curveFamily EccFamily) *KeyType
	EccPublicKey(curveFamily EccFamily) *KeyType
	DhKeyPair(groupFamily DhFamily) *KeyType
	DhPublicKey(groupFamily DhFamily) *KeyType
}

type keyTypeFactory struct{}

func NewKeyType() KeyTypeFactory {
	return &keyTypeFactory{}
}

func (a *keyTypeFactory) RawData() *KeyType {
	return &KeyType{
		variant: &KeyTypeRawData{},
	}
}

func (a *keyTypeFactory) Hmac() *KeyType {
	return &KeyType{variant: &KeyTypeHmac{}}
}
func (a *keyTypeFactory) Derive() *KeyType {
	return &KeyType{variant: &KeyTypeDerive{}}
}
func (a *keyTypeFactory) Aes() *KeyType {
	return &KeyType{variant: &KeyTypeAes{}}
}
func (a *keyTypeFactory) Des() *KeyType {
	return &KeyType{variant: &KeyTypeDes{}}
}
func (a *keyTypeFactory) Camellia() *KeyType {
	return &KeyType{variant: &KeyTypeCamellia{}}
}
func (a *keyTypeFactory) Arc4() *KeyType {
	return &KeyType{variant: &KeyTypeArc4{}}
}
func (a *keyTypeFactory) Chacha20() *KeyType {
	return &KeyType{variant: &KeyTypeChacha20{}}
}
func (a *keyTypeFactory) RsaPublicKey() *KeyType {
	return &KeyType{variant: &KeyTypeRsaPublicKey{}}
}
func (a *keyTypeFactory) RsaKeyPair() *KeyType {
	return &KeyType{variant: &KeyTypeRsaKeyPair{}}
}
func (a *keyTypeFactory) EccKeyPair(curveFamily EccFamily) *KeyType {
	return &KeyType{variant: &KeyTypeEccKeyPair{
		CurveFamily: curveFamily,
	}}
}
func (a *keyTypeFactory) EccPublicKey(curveFamily EccFamily) *KeyType {
	return &KeyType{variant: &KeyTypeEccPublicKey{
		CurveFamily: curveFamily,
	}}
}
func (a *keyTypeFactory) DhKeyPair(groupFamily DhFamily) *KeyType {
	return &KeyType{variant: &KeyTypeDhKeyPair{
		GroupFamily: groupFamily,
	}}
}
func (a *keyTypeFactory) DhPublicKey(groupFamily DhFamily) *KeyType {
	return &KeyType{variant: &KeyTypeDhPublicKey{
		GroupFamily: groupFamily,
	}}
}

type KeyType struct {
	// Types that are assignable to variant:
	//	*KeyType_RawData
	//	*KeyType_Hmac
	//	*KeyType_Derive
	//	*KeyType_Aes
	//	*KeyType_Des
	//	*KeyType_Camellia
	//	*KeyType_Arc4
	//	*KeyType_Chacha20
	//	*KeyType_RsaPublicKey
	//	*KeyType_RsaKeyPair
	//	*KeyType_EccKeyPair
	//	*KeyType_EccPublicKey
	//	*KeyType_DhKeyPair
	//	*KeyType_DhPublicKey
	variant keyTypeVariant
}

type keyTypeVariant interface {
	isKeyTypeVariant()
	toWireInterface() interface{}
}

func (k *KeyType) ToWireInterface() interface{} {
	if k.variant == nil {
		return nil
	}
	return k.variant.toWireInterface()
}

type EccFamily int32

const (
	KeyTypeECCFAMILYNONE EccFamily = 0 // This default variant should not be used.
	KeyTypeSECPK1        EccFamily = 1
	KeyTypeSECPR1        EccFamily = 2
	// Deprecated: Do not use.
	KeyTypeSECPR2 EccFamily = 3
	KeyTypeSECTK1 EccFamily = 4 // DEPRECATED for sect163k1 curve
	KeyTypeSECTR1 EccFamily = 5 // DEPRECATED for sect163r1 curve
	// Deprecated: Do not use.
	KeyTypeSECTR2       EccFamily = 6
	KeyTypeBRAINPOOLPR1 EccFamily = 7 // DEPRECATED for brainpoolP160r1 curve
	KeyTypeFRP          EccFamily = 8
	KeyTypeMONTGOMERY   EccFamily = 9
)

type DhFamily int32

const (
	KeyTypeRFC7919 DhFamily = 0
)

type KeyTypeRawData struct{}

func (k KeyTypeRawData) isKeyTypeVariant() {}
func (k *KeyTypeRawData) toWireInterface() interface{} {
	return &psakeyattributes.KeyType{
		Variant: &psakeyattributes.KeyType_RawData_{},
	}
}

type KeyTypeHmac struct{}

func (k KeyTypeHmac) isKeyTypeVariant() {}

func (k *KeyTypeHmac) toWireInterface() interface{} {
	return &psakeyattributes.KeyType{
		Variant: &psakeyattributes.KeyType_Hmac_{},
	}
}

type KeyTypeDerive struct{}

func (k KeyTypeDerive) isKeyTypeVariant() {}

func (k *KeyTypeDerive) toWireInterface() interface{} {
	return &psakeyattributes.KeyType{
		Variant: &psakeyattributes.KeyType_Derive_{},
	}
}

type KeyTypeAes struct{}

func (k KeyTypeAes) isKeyTypeVariant() {}

func (k *KeyTypeAes) toWireInterface() interface{} {
	return &psakeyattributes.KeyType{
		Variant: &psakeyattributes.KeyType_Aes_{},
	}
}

type KeyTypeDes struct{}

func (k KeyTypeDes) isKeyTypeVariant() {}

func (k *KeyTypeDes) toWireInterface() interface{} {
	return &psakeyattributes.KeyType{
		Variant: &psakeyattributes.KeyType_Des_{},
	}
}

type KeyTypeCamellia struct{}

func (k KeyTypeCamellia) isKeyTypeVariant() {}

func (k *KeyTypeCamellia) toWireInterface() interface{} {
	return &psakeyattributes.KeyType{
		Variant: &psakeyattributes.KeyType_Camellia_{},
	}
}

type KeyTypeArc4 struct{}

func (k KeyTypeArc4) isKeyTypeVariant() {}

func (k *KeyTypeArc4) toWireInterface() interface{} {
	return &psakeyattributes.KeyType{
		Variant: &psakeyattributes.KeyType_Arc4_{},
	}
}

type KeyTypeChacha20 struct{}

func (k KeyTypeChacha20) isKeyTypeVariant() {}

func (k *KeyTypeChacha20) toWireInterface() interface{} {
	return &psakeyattributes.KeyType{
		Variant: &psakeyattributes.KeyType_Chacha20_{},
	}
}

type KeyTypeRsaPublicKey struct{}

func (k KeyTypeRsaPublicKey) isKeyTypeVariant() {}

func (k *KeyTypeRsaPublicKey) toWireInterface() interface{} {
	return &psakeyattributes.KeyType{
		Variant: &psakeyattributes.KeyType_RsaPublicKey_{},
	}
}

type KeyTypeRsaKeyPair struct{}

func (k KeyTypeRsaKeyPair) isKeyTypeVariant() {}

func (k *KeyTypeRsaKeyPair) toWireInterface() interface{} {
	return &psakeyattributes.KeyType{
		Variant: &psakeyattributes.KeyType_RsaKeyPair_{},
	}
}

type KeyTypeEccKeyPair struct {
	CurveFamily EccFamily
}

func (k KeyTypeEccKeyPair) isKeyTypeVariant() {}

func (k *KeyTypeEccKeyPair) toWireInterface() interface{} {
	return &psakeyattributes.KeyType{
		Variant: &psakeyattributes.KeyType_EccKeyPair_{},
	}
}

type KeyTypeEccPublicKey struct {
	CurveFamily EccFamily
}

func (k KeyTypeEccPublicKey) isKeyTypeVariant() {}

func (k *KeyTypeEccPublicKey) toWireInterface() interface{} {
	return &psakeyattributes.KeyType{
		Variant: &psakeyattributes.KeyType_EccPublicKey_{},
	}
}

type KeyTypeDhKeyPair struct {
	GroupFamily DhFamily
}

func (k KeyTypeDhKeyPair) isKeyTypeVariant() {}

func (k *KeyTypeDhKeyPair) toWireInterface() interface{} {
	return &psakeyattributes.KeyType{
		Variant: &psakeyattributes.KeyType_DhKeyPair_{},
	}
}

type KeyTypeDhPublicKey struct{ GroupFamily DhFamily }

func (k KeyTypeDhPublicKey) isKeyTypeVariant() {}

func (k *KeyTypeDhPublicKey) toWireInterface() interface{} {
	return &psakeyattributes.KeyType{
		Variant: &psakeyattributes.KeyType_DhPublicKey_{},
	}
}
