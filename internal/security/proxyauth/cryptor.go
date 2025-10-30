//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package proxyauth

import (
	"bytes"
	"context"
	"crypto/aes"
	"os"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/pkg/utils/crypto"
	cryptoInterfaces "github.com/edgexfoundry/edgex-go/internal/pkg/utils/crypto/interfaces"
	proxyauthContainer "github.com/edgexfoundry/edgex-go/internal/security/proxyauth/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

const (
	defaultAESKeyFile = "/res/insecure_aes.key"
	EnvAesKey         = "EDGEX_INSECURE_AES_KEY"
)

// make getenv and readFile overridable for unit tests
var (
	getenv   = os.Getenv
	readFile = os.ReadFile
)

func loadDefaultAESKey() ([]byte, error) {
	if envKey := getenv(EnvAesKey); envKey != "" {
		return []byte(envKey), nil
	}

	if keyData, err := readFile(defaultAESKeyFile); err == nil {
		return bytes.TrimSpace(keyData), nil
	} else {
		return nil, err
	}
}

func checkAESKeySize(key []byte) error {
	k := len(key)
	switch k {
	default:
		return aes.KeySizeError(k)
	case 16, 24, 32:
		return nil
	}
}

func createAESCryptor(dic *di.Container) (cryptoInterfaces.Crypto, error) {
	defaultKey, err := loadDefaultAESKey()
	if err != nil {
		return nil, err
	}
	if err := checkAESKeySize(defaultKey); err != nil {
		return nil, err
	}
	if secret.IsSecurityEnabled() {
		secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)
		return crypto.NewAESCryptorWithSecretProvider(secretProvider, defaultKey)
	}
	return crypto.NewAESCryptor(defaultKey), nil
}

// AESCryptorBootstrapHandler creates and registers the AES cryptor
func AESCryptorBootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	cryptor, err := createAESCryptor(dic)
	if err != nil {
		lc.Errorf("failed to create AES cryptor: %v", err)
		return false
	}

	dic.Update(di.ServiceConstructorMap{
		proxyauthContainer.CryptoInterfaceName: func(get di.Get) interface{} {
			return cryptor
		},
	})

	return true
}
