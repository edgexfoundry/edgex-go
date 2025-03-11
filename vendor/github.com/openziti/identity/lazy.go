package identity

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/michaelquigley/pfxlog"
	"sync"
)

var _ Identity = &LazyIdentity{}

// LazyIdentity will delay calling identity.LoadIdentity(config) till it is first accessed.
type LazyIdentity struct {
	Identity
	*Config
	loadOnce sync.Once
}

func (self *LazyIdentity) load() {
	self.loadOnce.Do(func() {
		var err error
		self.Identity, err = LoadIdentity(*self.Config)

		if err != nil {
			pfxlog.Logger().Fatalf("error during lazy load of identity: %v", err)
		}
	})
}

func (self *LazyIdentity) Cert() *tls.Certificate {
	self.load()
	return self.Identity.Cert()
}

func (self *LazyIdentity) ServerCert() []*tls.Certificate {
	self.load()
	return self.Identity.ServerCert()
}

func (self *LazyIdentity) CA() *x509.CertPool {
	self.load()
	return self.Identity.CA()
}

func (self *LazyIdentity) CaPool() *CaPool {
	self.load()
	return self.Identity.CaPool()
}

func (self *LazyIdentity) ServerTLSConfig() *tls.Config {
	self.load()
	return self.Identity.ServerTLSConfig()
}

func (self *LazyIdentity) ClientTLSConfig() *tls.Config {
	self.load()
	return self.Identity.ClientTLSConfig()
}

func (self *LazyIdentity) Reload() error {
	self.load()
	return self.Identity.Reload()
}

func (self *LazyIdentity) WatchFiles() error {
	self.load()
	return self.Identity.WatchFiles()
}

func (self *LazyIdentity) StopWatchingFiles() {
	self.load()
	self.Identity.StopWatchingFiles()
}

func (self *LazyIdentity) SetCert(pem string) error {
	self.load()
	return self.Identity.SetCert(pem)
}

func (self *LazyIdentity) SetServerCert(pem string) error {
	self.load()
	return self.Identity.SetServerCert(pem)
}

func (self *LazyIdentity) GetConfig() *Config {
	self.load()
	return self.Identity.GetConfig()
}
