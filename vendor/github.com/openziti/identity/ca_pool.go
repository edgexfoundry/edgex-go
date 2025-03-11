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
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"github.com/pkg/errors"
	"time"
)

type CaPool struct {
	roots         map[string]*x509.Certificate
	intermediates map[string]*x509.Certificate
}

func NewCaPool(certs []*x509.Certificate) *CaPool {
	result := &CaPool{
		roots:         make(map[string]*x509.Certificate),
		intermediates: make(map[string]*x509.Certificate),
	}

	for _, cert := range certs {
		_ = result.AddCa(cert)
	}
	return result
}

// AddCa adds a CA (root or intermediate) certificate to the current pool. It returns an error if the
// certificate is not CA.
func (self *CaPool) AddCa(cert *x509.Certificate) error {
	if cert == nil {
		return errors.New("cannot add a nil certificate")
	}
	if !cert.IsCA {
		return errors.New("x509 certificates does not have the CA flag set to true")
	}

	hash := hashCertSha256(cert)

	if IsRootCa(cert) {
		self.roots[hash] = cert
	} else {
		self.intermediates[hash] = cert
	}

	return nil
}

// Roots returns a copy of the slice of currently added roots
func (self *CaPool) Roots() []*x509.Certificate {
	rootsCopy := make([]*x509.Certificate, len(self.roots))

	i := 0
	for _, cert := range self.roots {
		rootsCopy[i] = cert
		i++
	}

	return rootsCopy
}

// Intermediates returns a copy of the slice of currently added intermediates
func (self *CaPool) Intermediates() []*x509.Certificate {
	intermediatesCopy := make([]*x509.Certificate, len(self.intermediates))

	i := 0
	for _, cert := range self.intermediates {
		intermediatesCopy[i] = cert
		i++
	}

	return intermediatesCopy
}

// hashCertSha256 returns the sha256 fingerprint of a certificate
func hashCertSha256(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.Raw)
	return hex.EncodeToString(hash[:])
}

// rootsAsMap returns the root CAs in a new map
func (self *CaPool) rootsAsMap() map[string]*x509.Certificate {
	rootsCopy := make(map[string]*x509.Certificate, len(self.roots))
	for k, v := range self.roots {

		rootsCopy[k] = v
	}
	return rootsCopy
}

// intermediatesAsMap returns the intermediate CAs in a new map
func (self *CaPool) intermediatesAsMap() map[string]*x509.Certificate {
	intermediatesCopy := make(map[string]*x509.Certificate, len(self.intermediates))
	for k, v := range self.intermediates {
		intermediatesCopy[k] = v
	}
	return intermediatesCopy
}

// allCasAsMap returns all root and intermediate CAs in a new map
func (self *CaPool) allCasAsMap() map[string]*x509.Certificate {
	allCasMap := self.intermediatesAsMap()
	roots := self.rootsAsMap()
	for k, v := range roots {
		allCasMap[k] = v
	}

	return allCasMap
}

// GetChainMinusRoot returns a chain from `cert` up to, but not including, the root CA if possible. If no cert is
// provided, nil is returned, if no chains is assembled the resulting chain will be the target cert only.
func (self *CaPool) GetChainMinusRoot(cert *x509.Certificate, additionalCerts ...*x509.Certificate) []*x509.Certificate {
	if cert == nil {
		return nil
	}

	chainCandidates := self.intermediatesAsMap()

	for _, curCert := range additionalCerts {
		hash := hashCertSha256(curCert)
		chainCandidates[hash] = curCert
	}

	return assembleChain(cert, chainCandidates)
}

// GetChain returns a chain from `cert` up and including the root CA if possible. If no cert is provided, nil is
// returned. If no chains is assembled the resulting chain will be the target cert only.
func (self *CaPool) GetChain(cert *x509.Certificate, additionalCerts ...*x509.Certificate) []*x509.Certificate {
	if cert == nil {
		return nil
	}

	chainCandidates := self.allCasAsMap()

	for _, curCert := range additionalCerts {
		hash := hashCertSha256(curCert)
		chainCandidates[hash] = curCert
	}

	return assembleChain(cert, chainCandidates)
}

// VerifyToRoot will obtain a chain and verify it to a root CA. This is similar to the requirements that
// OpenSSL has for TLS.
func (self *CaPool) VerifyToRoot(cert *x509.Certificate) ([][]*x509.Certificate, error) {
	if cert == nil {
		return nil, errors.New("cannot verify a nil certificate")
	}

	opts := x509.VerifyOptions{
		Intermediates: self.IntermediatesAsStdPool(),
		Roots:         self.RootsAsStdPool(),
		CurrentTime:   time.Now(),
	}

	return cert.Verify(opts)
}

// IntermediatesAsStdPool returns all intermediates in an *x509.CertPool. Useful for calling standard x509 package
// functions.
func (self *CaPool) IntermediatesAsStdPool() *x509.CertPool {
	pool := x509.NewCertPool()

	for _, cert := range self.intermediates {
		pool.AddCert(cert)
	}

	return pool
}

// RootsAsStdPool returns all intermediates in an *x509.CertPool. Useful for calling standard x509 package
// functions.
func (self *CaPool) RootsAsStdPool() *x509.CertPool {
	pool := x509.NewCertPool()

	for _, cert := range self.roots {
		pool.AddCert(cert)
	}

	return pool
}

// assembleChain starts at `startCert` and build the longest chain up through ancestor signing certs as it can from `chainCandidates`.
func assembleChain(startCert *x509.Certificate, chainCandidates map[string]*x509.Certificate) []*x509.Certificate {
	if startCert == nil {
		return nil
	}

	var result []*x509.Certificate
	result = append(result, startCert)

	curCert := startCert

	for {
		if parent := getParent(curCert, chainCandidates); parent != nil {
			result = append(result, parent)
			curCert = parent
		} else {
			return result
		}
	}
}

// getParent returns the direct signing parent of a certificate found in a supplied map of certs. The supplied map
// is altered to remove the signing parent if found.
func getParent(cert *x509.Certificate, candidates map[string]*x509.Certificate) *x509.Certificate {
	for hash, candidate := range candidates {
		//cheaply check distinguishing names, verify with signature checking similar to OpenSSL
		if cert.Issuer.String() == candidate.Subject.String() {
			if err := cert.CheckSignatureFrom(candidate); err == nil {
				delete(candidates, hash)
				return candidate
			}
		}
	}
	return nil
}

// IsRootCa returns true if a certificate is a root certificate (is a ca, distinguishing name match on subject/issuer, and is self-signed)
func IsRootCa(cert *x509.Certificate) bool {
	//checking done in highest to lowest cost: CA flag, distinguishing name matching, signature check
	return cert.IsCA && cert.Issuer.String() == cert.Subject.String() && cert.CheckSignatureFrom(cert) == nil
}
