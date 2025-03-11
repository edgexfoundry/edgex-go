/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package identity

import (
	"crypto"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/tlz"
	"github.com/openziti/identity/certtools"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	StorageFile = "file"
	StoragePem  = "pem"
)

type Identity interface {
	Cert() *tls.Certificate
	ServerCert() []*tls.Certificate
	CA() *x509.CertPool
	CaPool() *CaPool
	ServerTLSConfig() *tls.Config
	ClientTLSConfig() *tls.Config
	Reload() error

	WatchFiles() error
	StopWatchingFiles()

	SetCert(pem string) error
	SetServerCert(pem string) error

	GetConfig() *Config
}

var _ Identity = &ID{}

type ID struct {
	Config

	certLock sync.RWMutex

	cert        *tls.Certificate
	serverCert  []*tls.Certificate
	ca          *x509.CertPool
	caPool      *CaPool
	needsReload atomic.Bool
	closeNotify chan struct{}
	watchCount  atomic.Int32
}

func (id *ID) initCert(loadedCerts []*x509.Certificate) error {
	chain := loadedCerts

	if id.caPool != nil {
		chain = id.caPool.GetChainMinusRoot(loadedCerts[0], loadedCerts[1:]...)
	}

	id.cert.Certificate = make([][]byte, len(chain))
	for i, c := range chain {
		id.cert.Certificate[i] = c.Raw
	}
	id.cert.Leaf = chain[0]
	return nil
}

// SetCert persists a new PEM as the ID's client certificate.
func (id *ID) SetCert(pem string) error {
	if certUrl, err := parseAddr(id.Config.Cert); err != nil {
		return err
	} else {
		switch certUrl.Scheme {
		case StoragePem:
			id.Config.Cert = StoragePem + ":" + pem
			return fmt.Errorf("could not save client certificate, location scheme not supported for saving (%s):\n%s", id.Config.Cert, pem)
		case StorageFile, "":
			f, err := os.OpenFile(id.Config.Cert, os.O_RDWR, 0664)
			if err != nil {
				return fmt.Errorf("could not update client certificate [%s]: %v", id.Config.Cert, err)
			}

			defer func() { _ = f.Close() }()

			err = f.Truncate(0)

			if err != nil {
				return fmt.Errorf("could not truncate client certificate [%s]: %v", id.Config.Cert, err)
			}

			_, err = fmt.Fprint(f, pem)

			if err != nil {
				return fmt.Errorf("error writing new client certificate [%s]: %v", id.Config.Cert, err)
			}
		default:
			return fmt.Errorf("could not save client certificate, location scheme not supported (%s) or address not defined (%s):\n%s", certUrl.Scheme, id.Config.Cert, pem)
		}
	}

	return nil
}

// SetServerCert persists a new PEM as the ID's server certificate.
func (id *ID) SetServerCert(pem string) error {
	if certUrl, err := parseAddr(id.Config.ServerCert); err != nil {
		return err
	} else {
		switch certUrl.Scheme {
		case StoragePem:
			id.Config.ServerCert = StoragePem + ":" + pem
			return fmt.Errorf("could not save client certificate, location scheme not supported for saving (%s): \n %s", id.Config.Cert, pem)
		case StorageFile, "":
			f, err := os.OpenFile(id.Config.ServerCert, os.O_RDWR, 0664)
			if err != nil {
				return fmt.Errorf("could not update server certificate [%s]: %v", id.Config.ServerCert, err)
			}

			defer func() { _ = f.Close() }()

			err = f.Truncate(0)

			if err != nil {
				return fmt.Errorf("could not truncate server certificate [%s]: %v", id.Config.ServerCert, err)
			}

			_, err = fmt.Fprint(f, pem)

			if err != nil {
				return fmt.Errorf("error writing new server certificate [%s]: %v", id.Config.ServerCert, err)
			}
		default:
			return fmt.Errorf("could not save server certificate, location scheme not supported (%s) or address not defined (%s):\n%s", certUrl.Scheme, id.Config.ServerCert, pem)
		}
	}

	return nil
}

// Reload re-interprets the internal Config that was used to create this ID. This instance of the
// ID is updated with new client, server, and ca configuration. All tls.Config's generated
// from this ID will use the newly loaded values for new connections.
func (id *ID) Reload() error {
	id.certLock.Lock()
	defer id.certLock.Unlock()

	newId, err := LoadIdentity(id.Config)

	if err != nil {
		return fmt.Errorf("failed to reload identity: %v", err)
	}

	id.ca = newId.CA()
	id.cert = newId.Cert()
	id.serverCert = newId.ServerCert()
	id.caPool = newId.CaPool()

	return nil
}

// getFiles returns all configuration paths that point to files
func (id *ID) getFiles() []string {
	var files []string
	if path, ok := IsFile(id.Config.Cert); ok {
		files = append(files, filepath.Clean(path))
	}

	if path, ok := IsFile(id.Config.ServerCert); ok {
		files = append(files, filepath.Clean(path))
	}

	if path, ok := IsFile(id.Config.Key); ok {
		files = append(files, filepath.Clean(path))
	}

	if path, ok := IsFile(id.Config.ServerKey); ok {
		files = append(files, filepath.Clean(path))
	}

	for _, altServerCert := range id.Config.AltServerCerts {
		if path, ok := IsFile(altServerCert.ServerKey); ok {
			files = append(files, filepath.Clean(path))
		}

		if path, ok := IsFile(altServerCert.ServerCert); ok {
			files = append(files, filepath.Clean(path))
		}

	}

	return files
}

// Cert returns the ID's current client certificate that is used by all tls.Config's generated from it.
func (id *ID) Cert() *tls.Certificate {
	id.certLock.RLock()
	defer id.certLock.RUnlock()

	return id.cert
}

// ServerCert returns the ID's current server certificate that is used by all tls.Config's generated from it.
func (id *ID) ServerCert() []*tls.Certificate {
	id.certLock.RLock()
	defer id.certLock.RUnlock()

	return id.serverCert
}

// CA returns the ID's current CA certificate pool that is used by all tls.Config's generated from it.
func (id *ID) CA() *x509.CertPool {
	id.certLock.RLock()
	defer id.certLock.RUnlock()

	return id.ca
}

// CaPool returns the ID's current CA certificate pool that can be used to build cert chains
func (id *ID) CaPool() *CaPool {
	id.certLock.RLock()
	defer id.certLock.RUnlock()

	return id.caPool
}

// ServerTLSConfig returns a new tls.Config instance that will delegate server certificate lookup to the current ID.
// Calling Reload on the source ID will update which server certificate is used if the internal Config is altered
// by calling Config or if the values the Config points to are altered (i.e. file update).
//
// Generating multiple tls.Config's by calling this method will return tls.Config's that are all tied to this ID's
// Config.
func (id *ID) ServerTLSConfig() *tls.Config {
	if id.serverCert == nil {
		return nil
	}

	tlsConfig := &tls.Config{
		GetCertificate: id.GetServerCertificate,
		RootCAs:        id.ca,
		ClientAuth:     tls.RequireAnyClientCert,
		MinVersion:     tlz.GetMinTlsVersion(),
		CipherSuites:   tlz.GetCipherSuites(),
	}

	//for servers, CAs can be updated for new connections by intercepting
	//on new client connections via GetConfigForClient
	tlsConfig.GetConfigForClient = func(info *tls.ClientHelloInfo) (*tls.Config, error) {
		return id.GetConfigForClient(tlsConfig, info)
	}

	return tlsConfig
}

// ClientTLSConfig returns a new tls.Config instance that will delegate client certificate lookup to the current ID.
// Calling Reload on the source ID can update which client certificate is used if the internal Config is altered
// by calling Config or if the values the Config points to are altered (i.e. file update).
//
// Generating multiple tls.Config's by calling this method will return tls.Config's that are all tied to this ID's
// Config and client certificates.
func (id *ID) ClientTLSConfig() *tls.Config {
	var tlsConfig *tls.Config = nil

	if id.ca != nil {
		tlsConfig = &tls.Config{
			RootCAs: id.ca,
		}
	}
	if id.cert != nil {
		if tlsConfig == nil {
			tlsConfig = &tls.Config{}
		}

		tlsConfig.GetClientCertificate = func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return id.GetClientCertificate(tlsConfig, info)
		}
	}

	return tlsConfig
}

// GetServerCertificate is used to satisfy tls.Config's GetCertificate requirements.
// Allows server certificates to be updated after enrollment extensions without stopping
// listeners and disconnecting clients. New settings are used for all new incoming connection.
func (id *ID) GetServerCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	id.certLock.RLock()
	defer id.certLock.RUnlock()

	if len(id.serverCert) == 0 {
		return nil, fmt.Errorf("no certificates")
	}

	if len(id.serverCert) == 1 {
		return id.serverCert[0], nil
	}

	for _, cert := range id.serverCert {
		if err := hello.SupportsCertificate(cert); err == nil {
			return cert, nil
		}
	}

	return id.serverCert[0], nil
}

// GetClientCertificate is used to satisfy tls.Config's GetClientCertificate requirements.
// Allows client certificates to be updated after enrollment extensions without disconnecting
// the current client. New settings will be used on re-connect.
func (id *ID) GetClientCertificate(config *tls.Config, _ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
	id.certLock.RLock()
	defer id.certLock.RUnlock()

	//root cas updated here because during the client connection process on the client side
	//tls.Config.GetConfigForClient is not called
	config.RootCAs = id.ca

	return id.cert, nil
}

// GetConfig returns the internally stored copy of the Config that was used to create
// the ID. The returned Config can be used to create additional IDs but those IDs
// will not share the same Config.
func (id *ID) GetConfig() *Config {
	return &id.Config
}

// GetConfigForClient is used to satisfy tls.Config's GetConfigForClient requirements.
// Allows servers to have up-to-date CA chains after enrollment extension.
func (id *ID) GetConfigForClient(config *tls.Config, _ *tls.ClientHelloInfo) (*tls.Config, error) {
	config.RootCAs = id.ca
	return config, nil
}

// queueReload de-duplicates reload attempts within a 1s window.
func (id *ID) queueReload(closeNotify <-chan struct{}) {
	if newReload := id.needsReload.CompareAndSwap(false, true); newReload {
		go func() {
			select {
			case <-time.After(1 * time.Second):
				if stillNeedsReload := id.needsReload.CompareAndSwap(true, false); stillNeedsReload {
					logrus.Info("reloading identity configuration")
					if err := id.Reload(); err != nil {
						logrus.Errorf("could not reload identity configuration: %v", err)
					}
				}
			case <-closeNotify:
				return
			}
		}()
	}
}

func LoadIdentity(cfg Config) (Identity, error) {
	id := &ID{
		Config: cfg,
	}

	var err error

	var defaultKey crypto.PrivateKey
	if cfg.Key != "" {
		var err error
		defaultKey, err = LoadKey(cfg.Key)
		if err != nil {
			return nil, err
		}
	}

	// CA bundle is optional, but can be used to fill in the client cert chain
	if cfg.CA != "" {
		if id.ca, id.caPool, err = loadCABundle(cfg.CA); err != nil {
			return nil, err
		}
	}

	if cfg.Cert != "" {
		if defaultKey == nil {
			return nil, errors.New("no key specified for identity cert")
		}

		id.cert = &tls.Certificate{
			PrivateKey: defaultKey,
		}

		if idCert, err := LoadCert(cfg.Cert); err != nil {
			return nil, err
		} else {
			if err = id.initCert(idCert); err != nil {
				return nil, err
			}
		}
	}

	// Server Cert is optional
	if cfg.ServerCert != "" {
		if svrCert, err := LoadCert(cfg.ServerCert); err != nil {
			return nil, err
		} else {
			var serverKey crypto.PrivateKey
			if cfg.ServerKey != "" {
				serverKey, err = LoadKey(cfg.ServerKey)
				if err != nil {
					return nil, err
				}
			} else if defaultKey != nil {
				serverKey = defaultKey
			} else {
				return nil, errors.New("no corresponding key specified for identity server_cert")
			}

			chains, err := AssembleServerChains(svrCert, nil)

			if err != nil {
				return nil, err
			}

			if strings.EqualFold("true", os.Getenv("ZT_DEBUG_CERTS")) {
				log := pfxlog.Logger()
				for i, chain := range chains {
					for j, cert := range chain {
						log.Infof("server cert [%v.%v] cn=%v isca=%v dns=%v ips=%v uris=%v", i, j,
							cert.Subject.CommonName, cert.IsCA, cert.DNSNames, cert.IPAddresses, cert.URIs)
					}
				}
			}

			tlsCerts := ChainsToTlsCerts(chains, serverKey)
			id.serverCert = append(id.serverCert, tlsCerts...)
		}
	}

	// Alt Server Cert is optional
	for _, altCert := range cfg.AltServerCerts {
		if svrCert, err := LoadCert(altCert.ServerCert); err != nil {
			return nil, err
		} else {
			var serverKey crypto.PrivateKey
			if altCert.ServerKey != "" {
				serverKey, err = LoadKey(altCert.ServerKey)
				if err != nil {
					return nil, err
				}
			} else if defaultKey != nil {
				serverKey = defaultKey
			} else {
				return nil, errors.New("no key specified for identity alternate server cert")
			}

			chains, err := AssembleServerChains(svrCert, nil)

			if err != nil {
				return nil, err
			}

			tlsCerts := ChainsToTlsCerts(chains, serverKey)
			id.serverCert = append(id.serverCert, tlsCerts...)
		}
	}

	return id, nil
}

// getUniqueCerts will return a slice of unique certificates from the given slice
func getUniqueCerts(certs []*x509.Certificate) []*x509.Certificate {
	set := map[string]*x509.Certificate{}
	var keys []string

	for _, cert := range certs {
		hash := sha1.Sum(cert.Raw)
		fp := string(hash[:])
		if _, exists := set[fp]; !exists {
			set[fp] = cert
			keys = append(keys, fp)
		}
	}

	var result []*x509.Certificate
	for _, key := range keys {
		result = append(result, set[key])
	}
	return result
}

// getUniqueCas will return a slice of unique certificates that are CAs from the given slice
func getUniqueCas(certs []*x509.Certificate) []*x509.Certificate {
	set := map[string]*x509.Certificate{}
	var keys []string

	for _, cert := range certs {
		if cert.IsCA {
			hash := sha1.Sum(cert.Raw)
			fp := string(hash[:])
			if _, exists := set[fp]; !exists {
				set[fp] = cert
				keys = append(keys, fp)
			}
		}
	}

	var result []*x509.Certificate
	for _, key := range keys {
		result = append(result, set[key])
	}
	return result
}

// LoadKey will inspect the string property from an identity configuration and attempt to load a private key
// from there. The type of location is determined by a format with a type prefix followed by a colon. If no
// known type prefix is present, it is assumed the entire value is a file path.
//
// Support Formats:
// - `pem:<PEM>`
// - `file:<PATH>`
func LoadKey(keyAddr string) (crypto.PrivateKey, error) {
	if keyUrl, err := parseAddr(keyAddr); err != nil {
		return nil, err
	} else {

		switch keyUrl.Scheme {
		case StoragePem:
			return certtools.LoadPrivateKey([]byte(keyUrl.Opaque))
		case StorageFile, "":
			return certtools.GetKey(nil, keyUrl.Path, "")
		default:
			// engine key format: "{engine_id}:{engine_opts} see specific engine for supported options
			return certtools.GetKey(keyUrl, "", "")
			//return nil, fmt.Errorf("could not load key, location scheme not supported (%s) or address not defined (%s)", keyUrl.Scheme, keyAddr)
		}
	}
}

// LoadCert will inspect the string property from an identity configuration and attempt to load an array of *x509.Certificate
// from there. The type of location is determined by a format with a type prefix followed by a colon. If no known type prefix is
// present, it is assumed the entire value is a file path.
//
// Support Formats:
// - `pem:<PEM>`
// - `file:<PATH>`
func LoadCert(certAddr string) ([]*x509.Certificate, error) {
	if certUrl, err := parseAddr(certAddr); err != nil {
		return nil, err
	} else {
		switch certUrl.Scheme {
		case StoragePem:
			return certtools.LoadCert([]byte(certUrl.Opaque))
		case StorageFile, "":
			return certtools.LoadCertFromFile(certUrl.Path)
		default:
			return nil, fmt.Errorf("could not load cert, location scheme not supported (%s) or address not defined (%s)", certUrl.Scheme, certAddr)
		}
	}
}

// IsFile returns a file path from a given configuration value and true if the configuration value is a file.
// Otherwise, returns empty string and false.
func IsFile(configValue string) (string, bool) {
	configValue = strings.TrimSpace(configValue)
	if configValue == "" {
		return "", false
	}

	if certUrl, err := parseAddr(configValue); err != nil {
		return "", false
	} else if certUrl.Scheme == StorageFile {
		return certUrl.Path, true
	}

	return "", false
}

func loadCABundle(caAddr string) (*x509.CertPool, *CaPool, error) {
	if caUrl, err := parseAddr(caAddr); err != nil {
		return nil, nil, err
	} else {
		pool := x509.NewCertPool()
		var bundle []byte
		switch caUrl.Scheme {
		case StoragePem:
			bundle = []byte(caUrl.Opaque)

		case StorageFile, "":
			if bundle, err = os.ReadFile(caUrl.Path); err != nil {
				return nil, nil, err
			}

		default:
			return nil, nil, errors.Errorf("invalid cert location, unsupported scheme: '%v'", caAddr)
		}

		pool.AppendCertsFromPEM(bundle)

		certs, err := certtools.LoadCert(bundle)
		if err != nil {
			return nil, nil, err
		}

		caPool := NewCaPool(certs)

		return pool, caPool, nil
	}
}
