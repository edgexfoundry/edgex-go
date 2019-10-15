/*******************************************************************************
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
 *******************************************************************************/

package certificates

import (
	"crypto"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/security/secrets"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

type rootCertGenerator struct {
	logger logger.LoggingClient
	seed   secrets.CertificateSeed
	writer FileWriter
}

func (gen rootCertGenerator) Generate() (err error) {
	if gen.seed.NewCA {
		gen.logger.Debug("<Phase 1> Generating CA PKI materials")
		gen.logger.Debug("Generating Root CA key pair (sk,pk)")

		private, err := generatePrivateKey(gen.seed, gen.logger)
		if err != nil {
			return err
		}

		// Extract PK from RSA or EC generated SK
		public := private.(crypto.Signer).Public()
		if gen.seed.DumpKeys {
			dumpKeyPair(private, gen.logger)
			dumpKeyPair(public, gen.logger)
		}

		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
		if err != nil {
			return fmt.Errorf("failed to generate serial number: %s", err.Error())
		}

		spkiASN1, err := x509.MarshalPKIXPublicKey(public)
		if err != nil {
			return fmt.Errorf("failed to encode public key: %s", err.Error())
		}

		var spki struct {
			Algorithm        pkix.AlgorithmIdentifier
			SubjectPublicKey asn1.BitString
		}
		_, err = asn1.Unmarshal(spkiASN1, &spki)
		if err != nil {
			return fmt.Errorf("failed to decode public key: %s", err.Error())
		}

		skid := sha1.Sum(spki.SubjectPublicKey.Bytes)
		caCertTemplate := &x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				CommonName:         gen.seed.CAName,
				Organization:       []string{gen.seed.CAName},
				OrganizationalUnit: []string{gen.seed.CAOrg},
				Locality:           []string{gen.seed.CALocality},
				Province:           []string{gen.seed.CAState},
				Country:            []string{gen.seed.CACountry},
			},
			EmailAddresses:        []string{gen.seed.CAName + "@" + gen.seed.TLSDomain},
			SubjectKeyId:          skid[:],
			NotAfter:              time.Now().AddDate(10, 0, 0),
			NotBefore:             time.Now(),
			KeyUsage:              x509.KeyUsageCertSign,
			BasicConstraintsValid: true,
			IsCA:                  true,
			MaxPathLenZero:        true,
		}
		gen.logger.Debug("Generating Root CA certificate")
		caDER, err := x509.CreateCertificate(rand.Reader, caCertTemplate, caCertTemplate, public, private)
		if err != nil {
			return fmt.Errorf("failed to generate CA certificate (DER): %s", err.Error())
		}

		_, err = x509.ParseCertificate(caDER)
		if err != nil {
			return fmt.Errorf("failed to parse Root CA certificate: %s", err.Error())
		}

		gen.logger.Debug(fmt.Sprintf("Saving Root CA private key to PEM file: %s", gen.seed.CAKeyFile))
		skPKCS8, err := x509.MarshalPKCS8PrivateKey(private)
		if err != nil {
			return fmt.Errorf("failed to encode CA private key: %s", err.Error())
		}

		err = gen.writer.Write(gen.seed.CAKeyFile, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: skPKCS8}), 0400)
		if err != nil {
			return fmt.Errorf("failed to save CA private key: %s", err.Error())
		}

		gen.logger.Debug(fmt.Sprintf("Saving Root CA certificate to PEM file: %s", gen.seed.CACertFile))
		err = gen.writer.Write(gen.seed.CACertFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}), 0644)
		if err != nil {
			return fmt.Errorf("failed to save CA certificate: %s", err.Error())
		}

		gen.logger.Debug("New local Root CA successfully created!")
	}
	return
}
