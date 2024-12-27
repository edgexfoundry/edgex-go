//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"testing"

	cryptoMocks "github.com/edgexfoundry/edgex-go/internal/pkg/utils/crypto/interfaces/mocks"
	"github.com/edgexfoundry/edgex-go/internal/security/proxyauth/container"
	"github.com/edgexfoundry/edgex-go/internal/security/proxyauth/infrastructure/interfaces/mocks"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/stretchr/testify/require"
)

func mockDic() *di.Container {
	return di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return logger.NewMockClient()
		},
	})
}

func TestAddKey(t *testing.T) {
	dic := mockDic()

	validNewKey := "validNewKey"
	validIssuer := "testIssuer"
	validKeyData := models.KeyData{
		Type:   common.VerificationKeyType,
		Issuer: validIssuer,
		Key:    validNewKey,
	}
	validKeyName := validKeyData.Issuer + "/" + validKeyData.Type
	validEncryptedKey := "encryptedValidNewKey"

	validUpdateKey := "validUpdateKey"
	updateKeyData := models.KeyData{
		Type:   common.SigningKeyType,
		Issuer: "issuer2",
		Key:    validUpdateKey,
	}
	validUpdateKeyName := updateKeyData.Issuer + "/" + updateKeyData.Type
	validUpdateEncryptedKey := "encryptedValidUpdateKey"

	invalidKeyData := models.KeyData{
		Type:   "invalidKeyType",
		Issuer: "issuer2",
		Key:    validUpdateKey,
	}

	encryptFailedKey := "encryptFailedKey"
	encryptFailedKeyData := models.KeyData{
		Type:   common.SigningKeyType,
		Issuer: "issuer3",
		Key:    encryptFailedKey,
	}

	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("KeyExists", validKeyName).Return(false, nil)
	dbClientMock.On("AddKey", validKeyName, validEncryptedKey).Return(nil)
	dbClientMock.On("KeyExists", validUpdateKeyName).Return(true, nil)
	dbClientMock.On("UpdateKey", validUpdateKeyName, validUpdateEncryptedKey).Return(nil)

	cryptoMock := &cryptoMocks.Crypto{}
	cryptoMock.On("Encrypt", validKeyData.Key).Return(validEncryptedKey, nil)
	cryptoMock.On("Encrypt", updateKeyData.Key).Return(validUpdateEncryptedKey, nil)
	cryptoMock.On("Encrypt", encryptFailedKeyData.Key).Return("", errors.NewCommonEdgeX(errors.KindServerError, "failed to encrypt the key", nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		container.CryptoInterfaceName: func(get di.Get) interface{} {
			return cryptoMock
		},
	})

	tests := []struct {
		name          string
		keyData       models.KeyData
		errorExpected bool
		errKind       errors.ErrKind
	}{
		{"Valid - Add new verification key", validKeyData, false, ""},
		{"Valid - Update existing signing key", updateKeyData, false, ""},
		{"Invalid - Invalid key type", invalidKeyData, true, errors.KindContractInvalid},
		{"Invalid - Encryption Error", encryptFailedKeyData, true, errors.KindServerError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := AddKey(dic, test.keyData)
			if test.errorExpected {
				require.Error(t, err)
				require.Equal(t, test.errKind, errors.Kind(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVerificationKeyByIssuer(t *testing.T) {
	dic := mockDic()

	validIssuer := "issuer1"
	validEncryptedKey := "encryptedKey"
	expectedKeyName := validIssuer + "/" + common.VerificationKeyType
	expectedKeyData := dtos.KeyData{Issuer: validIssuer, Type: common.VerificationKeyType, Key: "decryptedKey"}

	invalidIssuer := "invalidIssuer"
	invalidKeyName := invalidIssuer + "/" + common.VerificationKeyType

	decryptErrIssuer := "decryptErrIssuer"
	decryptErrKeyName := decryptErrIssuer + "/" + common.VerificationKeyType
	decryptErrKey := "decryptErrKey"

	dbClientMock := &mocks.DBClient{}
	dbClientMock.On("ReadKeyContent", expectedKeyName).Return(validEncryptedKey, nil)
	dbClientMock.On("ReadKeyContent", invalidKeyName).Return("", errors.NewCommonEdgeX(errors.KindServerError, "read key error", nil))
	dbClientMock.On("ReadKeyContent", decryptErrKeyName).Return(decryptErrKey, nil)

	cryptoMock := &cryptoMocks.Crypto{}
	cryptoMock.On("Decrypt", validEncryptedKey).Return([]byte("decryptedKey"), nil)
	cryptoMock.On("Decrypt", decryptErrKey).Return([]byte{}, errors.NewCommonEdgeX(errors.KindServerError, "decrypt key error", nil))

	dic.Update(di.ServiceConstructorMap{
		container.DBClientInterfaceName: func(get di.Get) interface{} {
			return dbClientMock
		},
		container.CryptoInterfaceName: func(get di.Get) interface{} {
			return cryptoMock
		},
	})

	tests := []struct {
		name            string
		issuer          string
		expectedKeyData dtos.KeyData
		errorExpected   bool
		errKind         errors.ErrKind
	}{
		{"Valid - Valid key", validIssuer, expectedKeyData, false, ""},
		{"Invalid - Empty issuer", "", dtos.KeyData{}, true, errors.KindContractInvalid},
		{"Invalid - Key read error", invalidIssuer, dtos.KeyData{}, true, errors.KindServerError},
		{"Invalid - Decryption error", decryptErrIssuer, dtos.KeyData{}, true, errors.KindServerError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := VerificationKeyByIssuer(dic, test.issuer)
			if test.errorExpected {
				require.Error(t, err)
				require.Equal(t, test.errKind, errors.Kind(err))
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedKeyData, result)
			}
		})
	}
}
