// Copyright (C) 2025 IOTech Ltd

package proxyauth

import (
	"context"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/pkg/utils/crypto"
	cryptoInterfaces "github.com/edgexfoundry/edgex-go/internal/pkg/utils/crypto/interfaces"
	proxyauthContainer "github.com/edgexfoundry/edgex-go/internal/security/proxyauth/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

func createAESCryptor(dic *di.Container) (cryptoInterfaces.Crypto, error) {
	if secret.IsSecurityEnabled() {
		secretProvider := bootstrapContainer.SecretProviderExtFrom(dic.Get)
		return crypto.NewAESCryptorWithSecretProvider(secretProvider)
	}
	return crypto.NewAESCryptor(), nil
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
