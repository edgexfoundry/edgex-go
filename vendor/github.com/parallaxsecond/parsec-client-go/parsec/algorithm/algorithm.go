package algorithm

import (
	"fmt"
	"reflect"

	"github.com/parallaxsecond/parsec-client-go/interface/operations/psaalgorithm"
)

type toWire interface {
	ToWireInterface() interface{}
}

type algorithmVariant interface {
	toWire
	isAlgorithmVariant()
}

type Algorithm struct {
	// Types that are assignable to Variant:
	//	*Algorithm_None_
	//	*Algorithm_Hash_
	//	*Algorithm_Mac_
	//	*Algorithm_Cipher_
	//	*Algorithm_Aead_
	//	*Algorithm_AsymmetricSignature_
	//	*Algorithm_AsymmetricEncryption_
	//	*Algorithm_KeyAgreement_
	//	*Algorithm_KeyDerivation_

	variant algorithmVariant
}

func (a *Algorithm) ToWireInterface() interface{} {
	if a.variant == nil {
		return nil
	}
	return a.variant.ToWireInterface()
}

func (a *Algorithm) GetAsymmetricSignature() *AsymmetricSignatureAlgorithm {
	if sub, ok := a.variant.(*AsymmetricSignatureAlgorithm); ok {
		return sub
	}
	return nil
}

func (a *Algorithm) GetAsymmetricEncryption() *AsymmetricEncryptionAlgorithm {
	if sub, ok := a.variant.(*AsymmetricEncryptionAlgorithm); ok {
		return sub
	}
	return nil
}

func (a *Algorithm) GetCipher() *Cipher {
	if sub, ok := a.variant.(*Cipher); ok {
		return sub
	}
	return nil
}
func (a *Algorithm) GetAead() *AeadAlgorithm {
	if sub, ok := a.variant.(*AeadAlgorithm); ok {
		return sub
	}
	return nil
}

func (a *Algorithm) GetHash() *HashAlgorithm {
	if sub, ok := a.variant.(*HashAlgorithm); ok {
		return sub
	}
	return nil
}

func NewAlgorithmFromWireInterface(op interface{}) (*Algorithm, error) {
	var algvar algorithmVariant
	var err error
	wireAlg, ok := op.(*psaalgorithm.Algorithm)
	if !ok {
		return nil, fmt.Errorf("expected psaalgorithm.Algorithm, got %v", reflect.TypeOf(op))
	}
	if a := wireAlg.GetAsymmetricSignature(); a != nil {
		algvar, err = newAsymmetricSignatureFromWire(a)
	}
	if a := wireAlg.GetAsymmetricEncryption(); a != nil {
		algvar, err = newAsymmetricEncryptionFromWire(a)
	}
	if a := wireAlg.GetAead(); a != nil {
		algvar, err = newAeadFromWire(a)
	}
	if a := wireAlg.GetCipher(); a != psaalgorithm.Algorithm_CIPHER_NONE {
		algvar, err = newCipherFromWire(a)
	}
	if a := wireAlg.GetHash(); a != psaalgorithm.Algorithm_HASH_NONE {
		algvar, err = newHashFromWire(a)
	}
	if a := wireAlg.GetKeyAgreement(); a != nil {
		algvar, err = newKeyAgreementFromWire(a)
	}
	if a := wireAlg.GetKeyDerivation(); a != nil {
		algvar, err = newKeyDerivationFromWire(a)
	}
	if a := wireAlg.GetMac(); a != nil {
		algvar, err = newMacFromWire(a)
	}
	// Not doing none

	if err != nil {
		return nil, err
	}
	if algvar != nil {
		return &Algorithm{
			variant: algvar,
		}, nil
	}
	return nil, fmt.Errorf("unknown algorithm type %v", reflect.TypeOf(op))
}
