//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/security/proxyauth/container"
	proxyAuthUtils "github.com/edgexfoundry/edgex-go/internal/security/proxyauth/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

// The AddKey function accepts the new KeyData model from the controller function
// and then invokes AddKey function of infrastructure layer to add new user
func AddKey(dic *di.Container, keyData models.KeyData) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	cryptor := container.CryptoFrom(dic.Get)

	keyName := ""
	if len(keyData.Type) == 0 {
		keyData.Type = common.VerificationKeyType
	}
	switch keyData.Type {
	case common.VerificationKeyType:
		keyName = proxyAuthUtils.VerificationKeyName(keyData.Issuer)
	case common.SigningKeyType:
		keyName = proxyAuthUtils.SigningKeyName(keyData.Issuer)
	default:
		return errors.NewCommonEdgeX(
			errors.KindContractInvalid,
			fmt.Sprintf("key type should be one of the '%s' or '%s'", common.VerificationKeyType, common.SigningKeyType), nil)
	}

	encryptedKey, err := cryptor.Encrypt(keyData.Key)
	if err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "failed to encrypt the key", err)
	}

	exists, edgexErr := dbClient.KeyExists(keyName)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}
	if exists {
		err = dbClient.UpdateKey(keyName, encryptedKey)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	} else {
		err = dbClient.AddKey(keyName, encryptedKey)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	return nil
}

// VerificationKeyByIssuer returns the verification key by issuer
func VerificationKeyByIssuer(dic *di.Container, issuer string) (dtos.KeyData, errors.EdgeX) {
	if issuer == "" {
		return dtos.KeyData{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "issuer is empty", nil)
	}
	keyName := proxyAuthUtils.VerificationKeyName(issuer)
	dbClient := container.DBClientFrom(dic.Get)
	cryptor := container.CryptoFrom(dic.Get)

	keyData, err := dbClient.ReadKeyContent(keyName)
	if err != nil {
		return dtos.KeyData{}, errors.NewCommonEdgeXWrapper(err)
	}

	decryptedKey, err := cryptor.Decrypt(keyData)
	if err != nil {
		return dtos.KeyData{}, errors.NewCommonEdgeX(errors.Kind(err), "failed to decrypt the key", err)
	}

	return dtos.KeyData{Issuer: issuer, Type: common.VerificationKeyType, Key: string(decryptedKey)}, nil
}
