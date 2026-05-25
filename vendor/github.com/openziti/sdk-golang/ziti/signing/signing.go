package signing

import (
	"crypto"
	"crypto/dsa" //nolint:staticcheck
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"github.com/google/uuid"
	"github.com/michaelquigley/pfxlog"
	"github.com/pkg/errors"
	"math/big"
	"reflect"
)

const (
	Format1Rsa   byte = 1
	Format2Dsa   byte = 2
	Format3Ecdsa byte = 3
)

func AssertIdentityWithSecret(privateKey interface{}) ([]byte, error) {
	var result []byte
	nonceUUID := uuid.New()
	nonce := nonceUUID[:]
	hashF := crypto.SHA512.New()
	_, _ = hashF.Write(nonce)
	nonce = hashF.Sum(nil)

	if rsaPk, ok := privateKey.(*rsa.PrivateKey); ok {
		result = append(result, Format1Rsa)
		result = appendSizedSlice(result, nonce)
		signature, err := rsa.SignPSS(rand.Reader, rsaPk, crypto.SHA512, nonce, nil)
		if err != nil {
			return nil, err
		}
		result = appendSizedSlice(result, signature)
		return result, nil
	}

	if dsaPk, ok := privateKey.(*dsa.PrivateKey); ok {
		result = append(result, Format2Dsa)
		result = appendSizedSlice(result, nonce)
		r, s, err := dsa.Sign(rand.Reader, dsaPk, nonce[:])
		if err != nil {
			return nil, err
		}
		result = appendSizedSlice(result, r.Bytes())
		result = appendSizedSlice(result, s.Bytes())
		return result, nil
	}

	if ecdsaPk, ok := privateKey.(*ecdsa.PrivateKey); ok {
		result = append(result, Format3Ecdsa)
		result = appendSizedSlice(result, nonce)
		r, s, err := ecdsa.Sign(rand.Reader, ecdsaPk, nonce[:])
		if err != nil {
			return nil, err
		}
		result = appendSizedSlice(result, r.Bytes())
		result = appendSizedSlice(result, s.Bytes())
		return result, nil
	}

	return nil, errors.Errorf("unhandled private key type %v", reflect.TypeOf(privateKey))
}

func GetVerifier(val []byte) (Verifier, error) {
	if len(val) < 1 {
		return nil, errors.Errorf("can't verify empty identity secret")
	}

	secretType := val[0]
	val = val[1:]

	nonce, val, err := consumeBytesValue("nonce", val)
	if err != nil {
		return nil, err
	}

	if secretType == Format1Rsa {
		signature, val, err := consumeBytesValue("signature", val)
		if err != nil {
			return nil, err
		}
		if len(val) != 0 {
			return nil, errors.Errorf("encoding error: still %v unconsumed bytes remaining in identity secret", len(val))
		}
		return &rsaVerifier{
			nonce:     nonce,
			signature: signature,
		}, nil
	}

	if secretType == Format2Dsa {
		rBytes, val, err := consumeBytesValue("dsa signature r component", val)
		if err != nil {
			return nil, err
		}
		sBytes, val, err := consumeBytesValue("dsa signature s component", val)
		if err != nil {
			return nil, err
		}
		if len(val) != 0 {
			return nil, errors.Errorf("encoding error: still %v unconsumed bytes remaining in identity secret", len(val))
		}

		r := big.NewInt(0).SetBytes(rBytes)
		s := big.NewInt(0).SetBytes(sBytes)

		return &dsaVerifier{
			nonce: nonce,
			r:     r,
			s:     s,
		}, nil
	}

	if secretType == Format3Ecdsa {
		rBytes, val, err := consumeBytesValue("ecdsa signature r component", val)
		if err != nil {
			return nil, err
		}
		sBytes, val, err := consumeBytesValue("ecdsa signature s component", val)
		if err != nil {
			return nil, err
		}
		if len(val) != 0 {
			return nil, errors.Errorf("encoding error: still %v unconsumed bytes remaining in identity secret", len(val))
		}

		r := big.NewInt(0).SetBytes(rBytes)
		s := big.NewInt(0).SetBytes(sBytes)

		return &ecdsaVerifier{
			nonce: nonce,
			r:     r,
			s:     s,
		}, nil
	}

	return nil, errors.Errorf("unsupported identity secret format %v", secretType)
}

type Verifier interface {
	Verify(publicKey interface{}) bool
}

type rsaVerifier struct {
	nonce     []byte
	signature []byte
}

func (v *rsaVerifier) Verify(publicKey interface{}) bool {
	if rsaPubKey, ok := publicKey.(*rsa.PublicKey); ok {
		return nil == rsa.VerifyPSS(rsaPubKey, crypto.SHA512, v.nonce, v.signature, nil)
	}
	pfxlog.Logger().Warnf("incorrect public key type. expected rsa.PublicKey, but was %v", reflect.TypeOf(publicKey))
	return false
}

type dsaVerifier struct {
	nonce []byte
	r     *big.Int
	s     *big.Int
}

func (v *dsaVerifier) Verify(publicKey interface{}) bool {
	if dsaPubKey, ok := publicKey.(*dsa.PublicKey); ok {
		return dsa.Verify(dsaPubKey, v.nonce, v.r, v.s)
	}
	pfxlog.Logger().Warnf("incorrect public key type. expected dsa.PublicKey, but was %v", reflect.TypeOf(publicKey))
	return false
}

type ecdsaVerifier struct {
	nonce []byte
	r     *big.Int
	s     *big.Int
}

func (v *ecdsaVerifier) Verify(publicKey interface{}) bool {
	if ecdsaPubKey, ok := publicKey.(*ecdsa.PublicKey); ok {
		return ecdsa.Verify(ecdsaPubKey, v.nonce, v.r, v.s)
	}
	pfxlog.Logger().Warnf("incorrect public key type. expected ecdsa.PublicKey, but was %v", reflect.TypeOf(publicKey))
	return false
}

func appendSizedSlice(slice []byte, val []byte) []byte {
	size := len(val)
	sizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBuf, uint32(size))
	slice = append(slice, sizeBuf...)
	return append(slice, val...)
}

func consumeBytesValue(name string, val []byte) ([]byte, []byte, error) {
	if len(val) < 4 {
		return nil, nil, errors.Errorf("encoding error, not enough data to read len(%v)", name)
	}
	size := binary.LittleEndian.Uint32(val[:4])
	val = val[4:]
	if len(val) < int(size) {
		return nil, nil, errors.Errorf("encoding error, not enough data to read %v, which has size %v. Only %v bytes available", name, size, len(val))
	}
	result := val[0:size]
	val = val[size:]
	return result, val, nil
}
