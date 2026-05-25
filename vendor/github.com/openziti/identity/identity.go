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
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/tlz"
	"github.com/openziti/identity/certtools"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	StorageFile = "file"
	StoragePem  = "pem"
)

type Identity interface {

	// Cert returns the current tls.Certificate linked to this identity's loaded certificates that is used for
	// client connections. The first certificate is always the cert` value loaded.
	Cert() *tls.Certificate

	// ServerCert returns the current tls.Certificate linked to this identity's loaded certificates that is used
	// to initiate server listeners. The first certificate is always the root `cert` or `serverCert` value loaded.
	// Alternative server certs follow.
	ServerCert() []*tls.Certificate

	// CA returns the identities currently loaded x509.CertPool
	CA() *x509.CertPool

	// CaPool returns a more friendly version of x509.CertPool, useful for inspection
	CaPool() *CaPool

	// ServerTLSConfig returns a TSL config linked to this identity and its configuration and certificates. Mutations
	// to the identity (i.e. reloads, updates) propagate to the returned tls.Config.
	ServerTLSConfig() *tls.Config

	// ClientTLSConfig returns a tls.Config linked to this identity and its configuration and certificates. Mutations
	// to the identity (i.e. reloads, updates) propagate to the returned tls.Config.
	ClientTLSConfig() *tls.Config

	// Reload reloads the identity. All changes are propagated to tls.Configs returned by ClientTLSConfig and ServerTLSConfig.
	Reload() error

	// WatchFiles causes this identity to automatically watch its identity file and all referenced files for updates.
	// File updates will call Reload.
	WatchFiles() error

	// StopWatchingFiles reversed WatchFiles.
	StopWatchingFiles()

	// IsCertSettable returns nil if the "cert" certificate storage supports writing, used before calling SetCert()
	IsCertSettable() error

	// SetCert updates the current client cert in use and saves it to the identity file.
	SetCert(pem string) error

	// IsServerCertSettable returns nil if the server certificate storage supports writing, used before calling SetServerCert()
	IsServerCertSettable() error

	// SetServerCert update the current server cert in use and saves it to the identity file.
	SetServerCert(pem string) error

	// GetConfig returns the config used to generate this identity.
	GetConfig() *Config

	// GetX509ActiveClientCertChain returns the client certificate in use as a slice in order of [Leaf->Supporting Certs]
	GetX509ActiveClientCertChain() []*x509.Certificate

	// GetX509ActiveServerCertChains returns an array of arrays of x509.Certificates. Each sub-array is a
	// chain ordered in [Leaf->Supporting Certs]. Each chain is either from the `server_cert` field if defined,
	// otherwise `cert`, and all alternative server certs.
	GetX509ActiveServerCertChains() [][]*x509.Certificate

	// GetX509IdentityServerCertChain returns only the chain from the `server_cert` (if defined) else the chain
	// from the `cert` field.
	GetX509IdentityServerCertChain() []*x509.Certificate

	// GetX509IdentityAltCertCertChains returns all of the chains from the `alt_server_cert` array
	GetX509IdentityAltCertCertChains() [][]*x509.Certificate

	// GetCaPool returns a clone of the current  CA pool
	GetCaPool() *CaPool

	// CheckServerCertSansForConflicts checks the current leaf server certificate for duplicate IP/DNS SANs, which
	// cause ambiguous SNI lookups. Returns nil if no errors.
	CheckServerCertSansForConflicts() []SanHostConflictError

	// ValidFor checks a hostname or IP against all available server certificates and their SANs.
	ValidFor(hostnameOrIp string) error
}

type SanHostConflictError struct {
	HostOrIp     string
	Certificates []*x509.Certificate
}

func (s SanHostConflictError) Error() string {
	var certSubjects []string
	for _, cert := range s.Certificates {
		certSubjects = append(certSubjects, cert.Subject.String())
	}

	return fmt.Sprintf("the hostname/ip %s is handled by more than one certificate and should only be handled by one as it is ambiguous on which to use, certificate subjects: %s", s.HostOrIp, strings.Join(certSubjects, ", "))
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

func (id *ID) GetX509IdentityServerCertChain() []*x509.Certificate {
	id.certLock.Lock()
	defer id.certLock.Unlock()

	if id.Config.ServerCert != "" {
		chain, _ := LoadCert(id.Config.ServerCert)
		return chain
	}

	if id.Config.Cert != "" {
		chain, _ := LoadCert(id.Config.Cert)
		return chain
	}

	return nil
}

func (id *ID) GetX509ActiveClientCertChain() []*x509.Certificate {
	tlsCert := id.Cert()

	if len(tlsCert.Certificate) == 0 {
		return nil
	}

	var chain []*x509.Certificate
	for _, certDer := range tlsCert.Certificate {
		cert, _ := x509.ParseCertificate(certDer)
		chain = append(chain, cert)
	}

	return chain
}

func (id *ID) GetX509IdentityAltCertCertChains() [][]*x509.Certificate {
	id.certLock.Lock()
	defer id.certLock.Unlock()

	var chains [][]*x509.Certificate
	for _, keyPair := range id.Config.AltServerCerts {
		chain, _ := LoadCert(keyPair.ServerCert)
		chains = append(chains, chain)
	}

	return chains
}

func (id *ID) GetX509ActiveServerCertChains() [][]*x509.Certificate {
	var result [][]*x509.Certificate

	for _, rawChain := range id.ServerCert() {
		var parsedChain []*x509.Certificate
		for _, curCert := range rawChain.Certificate {
			cert, _ := x509.ParseCertificate(curCert)
			parsedChain = append(parsedChain, cert)
		}

		result = append(result, parsedChain)
	}

	return result
}

func (id *ID) GetCaPool() *CaPool {
	id.certLock.Lock()
	defer id.certLock.Unlock()

	return id.caPool.Clone()
}

func (id *ID) CheckServerCertSansForConflicts() []SanHostConflictError {
	var sanErrors []SanHostConflictError
	hostnames := map[string][]*x509.Certificate{}
	ipAddresses := map[string][]*x509.Certificate{}

	chains := id.GetX509ActiveServerCertChains()
	var certs []*x509.Certificate

	for _, chain := range chains {
		if len(chain) != 0 {
			certs = append(certs, chain[0])
		}
	}

	for _, cert := range certs {
		for _, dnsName := range cert.DNSNames {
			hostnames[dnsName] = append(hostnames[dnsName], cert)
		}

		for _, ip := range cert.IPAddresses {
			ipAddresses[ip.String()] = append(ipAddresses[ip.String()], cert)
		}
	}

	for ip, ipCerts := range ipAddresses {
		if len(ipCerts) > 1 {
			sanErrors = append(sanErrors, SanHostConflictError{
				HostOrIp:     ip,
				Certificates: ipCerts,
			})
		}
	}

	for hostname, hostnameCerts := range hostnames {
		if len(hostnameCerts) > 1 {
			sanErrors = append(sanErrors, SanHostConflictError{
				HostOrIp:     hostname,
				Certificates: hostnameCerts,
			})
		}
	}

	return sanErrors
}

func (id *ID) GetX509CaPool() *CaPool {
	id.certLock.Lock()
	defer id.certLock.Unlock()
	return id.caPool.Clone()
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

func (id *ID) IsCertSettable() error {
	certUrl, err := parseAddr(id.Config.Cert)

	if err != nil {
		return err
	}

	switch certUrl.Scheme {
	case StoragePem:
		return errors.New("cannot save cert in pem storage format")
	case StorageFile, "":
		absPath, err := filepath.Abs(id.Config.Cert)

		if err != nil {
			return fmt.Errorf("cannot get absolute path for cert file %s: %w", id.Config.Cert, err)
		}

		f, err := os.OpenFile(absPath, os.O_RDWR, 0664)

		if err != nil {
			return fmt.Errorf("can not save cert certificate [%s] due to file error: %v", absPath, err)
		}
		defer func() { _ = f.Close() }()

		return nil
	}

	return fmt.Errorf("can not save cert certificate, location scheme not supported (%s) or address not defined (%s)", certUrl.Scheme, id.Config.Cert)
}

func (id *ID) IsServerCertSettable() error {
	certUrl, err := parseAddr(id.Config.ServerCert)

	if err != nil {
		return err
	}

	switch certUrl.Scheme {
	case StoragePem:
		return errors.New("cannot save server cert in pem storage format")
	case StorageFile, "":
		absPath, err := filepath.Abs(id.Config.ServerCert)

		if err != nil {
			return fmt.Errorf("cannot get absolute path for server cert file %s: %w", id.Config.ServerCert, err)
		}

		f, err := os.OpenFile(absPath, os.O_RDWR, 0664)

		if err != nil {
			return fmt.Errorf("can not save server certificate [%s] due to file error: %v", absPath, err)
		}
		defer func() { _ = f.Close() }()

		return nil
	}

	return fmt.Errorf("can not save server certificate, location scheme not supported (%s) or address not defined (%s)", certUrl.Scheme, id.Config.ServerCert)
}

// SetCert persists a new PEM as the ID's client certificate.
func (id *ID) SetCert(pemStr string) error {
	certUrl, err := parseAddr(id.Config.Cert)

	if err != nil {
		return err
	}

	switch certUrl.Scheme {
	case StoragePem:
		id.Config.Cert = StoragePem + ":" + pemStr
		return fmt.Errorf("could not save cert certificate, location scheme not supported for saving (%s):\n%s", id.Config.Cert, pemStr)
	case StorageFile, "":

		absPath, err := filepath.Abs(id.Config.Cert)

		if err != nil {
			return fmt.Errorf("cannot get absolute path for cert file %s: %w", id.Config.Cert, err)
		}

		f, err := os.OpenFile(absPath, os.O_RDWR, 0664)
		if err != nil {
			return fmt.Errorf("could not update cert certificate [%s]: %v", id.Config.Cert, err)
		}

		defer func() { _ = f.Close() }()

		err = f.Truncate(0)

		if err != nil {
			return fmt.Errorf("could not truncate cert certificate [%s]: %v", id.Config.Cert, err)
		}

		_, err = fmt.Fprint(f, pemStr)

		if err != nil {
			return fmt.Errorf("error writing new cert certificate [%s]: %v", id.Config.Cert, err)
		}
	default:
		return fmt.Errorf("could not save cert certificate, location scheme not supported (%s) or address not defined (%s):\n%s", certUrl.Scheme, id.Config.Cert, pemStr)
	}

	return nil
}

// SetServerCert persists a new PEM as the ID's server certificate.
func (id *ID) SetServerCert(pem string) error {
	certUrl, err := parseAddr(id.Config.ServerCert)
	if err != nil {
		return err
	}

	switch certUrl.Scheme {
	case StoragePem:
		id.Config.ServerCert = StoragePem + ":" + pem
		return fmt.Errorf("could not save server certificate, location scheme not supported for saving (%s): \n %s", id.Config.ServerKey, pem)
	case StorageFile, "":

		absPath, err := filepath.Abs(id.Config.ServerCert)

		if err != nil {
			return fmt.Errorf("cannot get absolute path for server cert file %s: %w", id.Config.ServerCert, err)
		}

		f, err := os.OpenFile(absPath, os.O_RDWR, 0664)
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
		MaxVersion:     tlz.GetMaxTlsVersion(),
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
			// engine key format: "{engine_id}:{engine_opts}" see specific engine for supported options
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
			//try to figure it out like the c-sdk

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

func (id *ID) ValidFor(hostnameOrIp string) error {
	return ValidFor(id, hostnameOrIp)
}

// Define base errors
var (
	// ErrInvalidAddressForIdentity is returned during ip/hostname SANs validation. It represents that the ip/hostname
	// is not present as a SAN in any available server certificates.
	ErrInvalidAddressForIdentity = errors.New("identity is not valid for provided host")
)

// AddressError is returned during ip/hostname SANs validation. It represents that the ip/hostname is not present as
// a SAN in any available server certificates.
type AddressError struct {
	BaseErr  error
	Host     string
	ValidFor []string
}

func (e *AddressError) Error() string {
	return fmt.Sprintf("%s: [%s]. is valid for: [%s]", e.BaseErr.Error(), e.Host, strings.Join(e.ValidFor, ", "))
}

func (e *AddressError) Unwrap() error {
	return e.BaseErr
}

// ValidFor checks if the identity is valid for the given address
func ValidFor(id Identity, hostnameOrIp string) error {
	var err error
	// Check server certificate
	for _, c := range id.ServerCert() {
		err = c.Leaf.VerifyHostname(hostnameOrIp)
		if err == nil {
			return nil
		}
	}

	// Check client certificate if server cert validation fails
	if err != nil && id.Cert() != nil && id.Cert().Leaf != nil {
		err = id.Cert().Leaf.VerifyHostname(hostnameOrIp)
	}

	if err != nil {
		return &AddressError{BaseErr: ErrInvalidAddressForIdentity, Host: hostnameOrIp, ValidFor: getUniqueAddresses(id)}
	}
	if len(id.ServerCert()) == 0 && id.Cert() == nil {
		return &AddressError{BaseErr: ErrInvalidAddressForIdentity, Host: hostnameOrIp}
	}

	return nil
}

// getUniqueAddresses extracts unique DNS names and IP addresses from the identity's certificates
func getUniqueAddresses(id Identity) []string {
	addresses := make(map[string]struct{})

	if certs := id.ServerCert(); len(certs) > 0 && certs[0].Leaf != nil {
		for _, dns := range certs[0].Leaf.DNSNames {
			addresses[dns] = struct{}{}
		}
		for _, ip := range certs[0].Leaf.IPAddresses {
			addresses[ip.String()] = struct{}{}
		}
	}

	if cert := id.Cert(); cert != nil && cert.Leaf != nil {
		for _, dns := range cert.Leaf.DNSNames {
			addresses[dns] = struct{}{}
		}
		for _, ip := range cert.Leaf.IPAddresses {
			addresses[ip.String()] = struct{}{}
		}
	}

	uniqueList := make([]string, 0, len(addresses))
	for addr := range addresses {
		uniqueList = append(uniqueList, addr)
	}
	sort.Strings(uniqueList) // Ensure consistent order, mostly for testing

	return uniqueList
}
