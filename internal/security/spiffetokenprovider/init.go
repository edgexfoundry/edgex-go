/*******************************************************************************
 * Copyright 2022 Intel Corporation
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package spiffetokenprovider

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/spiffetokenprovider/container"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/authtokenloader"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"
	"github.com/edgexfoundry/go-mod-secrets/v2/secrets"
)

const (
	redisSecretName                  = "redisdb"
	secretBasePath                   = "/v1/secret/edgex" // nolint:gosec
	edgexRedisBootstrapperServiceKey = "security-bootstrapper-redis"
)

type Bootstrap struct {
	insecureSkipVerify bool
	vaultInterval      int
	validKnownSecrets  map[string]bool
}

func NewBootstrap(insecureSkipVerify bool, vaultInterval int) *Bootstrap {
	return &Bootstrap{
		insecureSkipVerify: insecureSkipVerify,
		vaultInterval:      vaultInterval,
		validKnownSecrets:  map[string]bool{redisSecretName: true},
	}
}

func (b *Bootstrap) getSecretStoreClient(dic *di.Container) (secrets.SecretStoreClient, error) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)
	secretStoreConfig := configuration.SecretStore

	fileOpener := fileioperformer.NewDefaultFileIoPerformer()

	var httpCaller internal.HttpCaller
	if caFilePath := secretStoreConfig.RootCaCertPath; caFilePath != "" {
		lc.Info("using certificate verification for secret store connection")
		caReader, err := fileOpener.OpenFileReader(caFilePath, os.O_RDONLY, 0400)
		if err != nil {
			return nil, err
		}
		httpCaller = pkg.NewRequester(lc).WithTLS(caReader, secretStoreConfig.ServerName)
	} else {
		lc.Info("bypassing certificate verification for secret store connection")
		httpCaller = pkg.NewRequester(lc).Insecure()
	}

	clientConfig := types.SecretConfig{
		Type:     secretStoreConfig.Type,
		Protocol: secretStoreConfig.Protocol,
		Host:     secretStoreConfig.Host,
		Port:     secretStoreConfig.Port,
	}
	secretClient, err := secrets.NewSecretStoreClient(clientConfig, lc, httpCaller)
	if err != nil {
		return nil, err
	}

	return secretClient, nil
}

func (b *Bootstrap) getPrivilegedToken(dic *di.Container) (string, error) {

	tokenLoader := bootstrapContainer.AuthTokenLoaderFrom(dic.Get)
	if tokenLoader == nil {
		tokenLoader = authtokenloader.NewAuthTokenLoader(fileioperformer.NewDefaultFileIoPerformer())
	}

	// Reload token in case new token was created causing the auth error
	configuration := container.ConfigurationFrom(dic.Get)
	token, err := tokenLoader.Load(configuration.GetBootstrap().SecretStore.TokenFile)
	if err != nil {
		return "", err
	}

	return token, nil

}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func (b *Bootstrap) BootstrapHandler(ctx context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)

	secretStoreClient, err := b.getSecretStoreClient(dic)
	if err != nil {
		lc.Errorf("failed to create SecretStoreClient: %s", err.Error())
		return false
	}

	// Handle healthcheck endpoint
	http.HandleFunc("/api/v2/ping", func(w http.ResponseWriter, r *http.Request) {
		lc.Info("Request received for /api/v2/ping")
		_, _ = io.WriteString(w, "pong")
	})

	//
	// The SPIFFE token provider will take three parameters:
	//
	// An X.509 SVID used in mutual-auth TLS for the token provider and the service to cross-authenticate.
	//
	// The requested service key. If blank, the service key will default to the service name encoded in the SVID.
	// If the service name follows the pattern device-(name), then the service key must follow the format
	// device-(name) or device-name-*. If the service name is app-service-configurable,
	// then the service key must follow the format app-*. (This is an accommodation for the Unix workload
	// attester not being able to distingish workloads that are launched using the same executable binary.
	// Custom app services that support multiple instances won't be supported unless they name the executable
	// the same as the standard app service binary or modify this logic.)
	//
	// A list of "known secret" identifiers that will allow new services to request database passwords or other
	// "known secrets" to be seeded into their service's partition in the secret store.
	//

	// Handle gettoken endpoint
	http.HandleFunc("/api/v2/gettoken", func(w http.ResponseWriter, r *http.Request) {
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)

		// Have to read the connection to finish TLS verification
		if err := r.ParseForm(); err != nil {
			lc.Errorf("Could not parse form: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		serviceKey := r.Form.Get("service_key")
		// could get multiple secret names from the form posted. e.g.: redisdb,mongodb,foo
		// for the known secrets we currently only supports redisdb
		// Due to the service key from the form only sees one service for the current use case
		knownSecretNames := r.Form["known_secret_names"]
		rawTokenOnly, err := strconv.ParseBool(r.Form.Get("raw_token"))
		if err != nil {
			lc.Errorf("Could not parse bool '%s': %v", r.Form.Get("raw_token"), err)
		}
		peerSVID := ""

	iterateCertificates:
		for _, cert := range r.TLS.PeerCertificates {
			for _, uri := range cert.URIs {
				// First certificate in verified chain has URI list,
				// one of which is the SVID containing the EdgeX service key
				if strings.HasPrefix(uri.String(), "spiffe://") {
					peerSVID = uri.String()
					break iterateCertificates
				}
			}
		}

		fmt.Println("PeerSVID:", peerSVID)
		// TODO: verify the prefix with what we expect like spiffe://edgexfoundry.org/service/*
		// regex: regexp.MustCompile(`(?m)^(?:spiffe://)(.)*/service/(.)*$`)
		spiffeServiceKey := strings.Replace(peerSVID,
			fmt.Sprintf("spiffe://%s/service/", configuration.Spiffe.TrustDomain.HostName), "", 1)

		privilegedToken, err := b.getPrivilegedToken(dic)
		if err != nil {
			lc.Errorf("failed to load secret store token: %v", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		vaultTokenResponse, err := makeToken(spiffeServiceKey, privilegedToken, secretStoreClient, lc)
		if err != nil {
			lc.Errorf("failed create secret store token: %v", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if err := b.seedKnownSecrets(ctx, lc, configuration.SecretStore, knownSecretNames, serviceKey, privilegedToken); err != nil {
			lc.Errorf("failed to seed known secrets: %v", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// TODO: Use above logic to compare SPIFFE ID with service key in form
		retText := fmt.Sprintf("svid=%s; svid_service=%s; service_key=%s; known_secret_names=%s, raw_token=%v\n",
			peerSVID, spiffeServiceKey, serviceKey, knownSecretNames, rawTokenOnly)
		lc.Info(retText)

		// Write resulting token
		if rawTokenOnly {
			w.Header().Add("Content-Type", "text/plain")
			rawToken := ((vaultTokenResponse).(map[string]interface{})["auth"]).(map[string]interface{})["client_token"].(string)
			if _, err := io.WriteString(w, rawToken); err != nil {
				lc.Errorf("Could not write output: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			w.Header().Add("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(vaultTokenResponse); err != nil {
				lc.Errorf("failed to write token response: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

	})

	// Get x.509 SVID from specified workload API socket
	socketPath := configuration.Spiffe.Endpoint.Socket
	if !strings.HasPrefix(socketPath, "unix://") {
		socketPath = "unix://" + socketPath
	}

	// command-line equivalent: spire-agent api fetch x509 -socketPath xxxx
	source, err := workloadapi.NewX509Source(ctx, workloadapi.WithClientOptions(workloadapi.WithAddr(socketPath)))
	if err != nil {
		lc.Errorf("Unable to create X509Source: %v", err)
		return false
	}
	defer source.Close()

	// This service can only be connected to by the local trust domain

	trustDomainString := os.Getenv("SPIFFE_TRUST_DOMAIN")
	if trustDomainString == "" {
		lc.Error("SPIFFE_TRUST_DOMAIN environment variable must be set")
		return false
	}
	td, err := spiffeid.TrustDomainFromString(trustDomainString)
	if err != nil {
		lc.Error("Could not get SPIFFE trust domain from string '%s': %v", trustDomainString, err)
		return false
	}

	// Create a `tls.Config` to allow mTLS connections, and verify that presented certificate has SPIFFE ID `spiffe://example.org/client`
	tlsConfig := tlsconfig.MTLSServerConfig(source, source, tlsconfig.AuthorizeMemberOf(td))
	tlsConfig.MinVersion = tls.VersionTLS13
	tlsConfig.CurvePreferences = []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256}
	tlsConfig.PreferServerCipherSuites = true

	serverAddress := ":" + strconv.Itoa(configuration.GetBootstrap().Service.Port)
	server := &http.Server{
		Addr:      serverAddress,
		TLSConfig: tlsConfig,
	}

	if err := server.ListenAndServeTLS("", ""); err != nil {
		lc.Errorf("Error on serve: %v", err)
	}

	return false
}

// seedKnownSecrets seeds or copies the known secrets from the existing service (e.g. security-bootstrapper-redis)
// to the requested new service that also uses the same known secrets
func (b *Bootstrap) seedKnownSecrets(ctx context.Context, lc logger.LoggingClient,
	ssConfig bootstrapConfig.SecretStoreInfo,
	knownSecretNames []string, serviceKey string, privilegedToken string) error {

	// to see if we can find redisdb as part of known secret name since that is the known secret we can support now
	found := false
	for _, secretName := range knownSecretNames {
		_, valid := b.validKnownSecrets[secretName]
		if valid {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot find secret name from validKnownSecrets")
	}

	// copy from security-bootstrapper-redis: /v1/secret/edgex/security-bootstrapper-redis/redisdb
	// to /v1/secret/edgex/<service_key>/redisdb using secret client's APIs

	secretConfig := types.SecretConfig{
		Type:           ssConfig.Type,
		Host:           ssConfig.Host,
		Port:           ssConfig.Port,
		Path:           secretBasePath, // make sure path is like /v1/edgex/secrets/ in global area
		SecretsFile:    ssConfig.SecretsFile,
		Protocol:       ssConfig.Protocol,
		Namespace:      ssConfig.Namespace,
		RootCaCertPath: ssConfig.RootCaCertPath,
		ServerName:     ssConfig.ServerName,
		Authentication: ssConfig.Authentication,
	}

	secretConfig.Authentication.AuthToken = privilegedToken

	secretClient, err := secrets.NewSecretsClient(ctx, secretConfig, lc, func(string) (string, bool) {
		return privilegedToken, true
	})
	if err != nil {
		return fmt.Errorf("found error on getting secretClient: %v", err)
	}

	// copy known secrets redisdb from redis-bootstrapper to the requested service with serviceKey
	secrets, err := secretClient.GetSecrets(fmt.Sprintf("/%s/%s", edgexRedisBootstrapperServiceKey, redisSecretName))
	if err != nil {
		return fmt.Errorf("found error on getting secrets: %v", err)
	}

	err = secretClient.StoreSecrets(fmt.Sprintf("/%s/%s", serviceKey, redisSecretName), secrets)
	if err != nil {
		return fmt.Errorf("found error on storing secrets: %v", err)
	}

	return nil
}
