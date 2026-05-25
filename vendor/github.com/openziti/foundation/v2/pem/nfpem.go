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

// Package nfpem provides convenience functions for dealing with PEM encoded x509 data.
package nfpem

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// DecodeAll accepts a byte array of PEM encoded data returns PEM blocks of data.
// The blocks will be in the order that they are provided in the original bytes.
func DecodeAll(pemBytes []byte) []*pem.Block {
	var blocks []*pem.Block
	if len(pemBytes) < 1 {
		return blocks
	}
	b, rest := pem.Decode(pemBytes)

	for b != nil {
		blocks = append(blocks, b)
		b, rest = pem.Decode(rest)
	}
	return blocks
}

// EncodeToString returns PEM encoded data in the form of a string generated
// from the supplied certificate.
func EncodeToString(cert *x509.Certificate) string {
	return string(EncodeToBytes(cert))
}

// EncodeToBytes returns PEM encoded data in the form of a string generated
// from the supplied certificate.
func EncodeToBytes(cert *x509.Certificate) []byte {
	result := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	return result
}

// PemToX509 accepts a PEM string and returns an array of x509.Certificate. Any blocks that
// cannot be parsed as certificates are discard. Certificate are returned to the
// order they are encountered in the PEM string.
//
// Deprecated: Use PemStringToCertificates
func PemToX509(pem string) []*x509.Certificate {
	return PemStringToCertificates(pem)
}

// PemBytesToCertificates accepts PEM bytes and returns an array of x509.Certificate. Any blocks that
// cannot be parsed as a x509.Certificate is discard. Certificate are returned to the
// order they are encountered in the PEM string.
func PemBytesToCertificates(pem []byte) []*x509.Certificate {
	pemBytes := pem
	certs := make([]*x509.Certificate, 0)
	for _, block := range DecodeAll(pemBytes) {
		xcerts, err := x509.ParseCertificate(block.Bytes)
		if err == nil && xcerts != nil {
			certs = append(certs, xcerts)
		}
	}
	return certs
}

// PemStringToCertificates accepts a PEM string and returns an array of x509.Certificate. Any blocks that
// cannot be parsed as certificates are discard. Certificate are returned to the
// order they are encountered in the PEM string.
func PemStringToCertificates(pem string) []*x509.Certificate {
	return PemBytesToCertificates([]byte(pem))
}

// FingerprintFromPem returns the fingerprint of the first certificate encountered in a pem string
// Deprecated: Use FingerprintFromPemString or FingerprintFromPemBytes
func FingerprintFromPem(pem string) string {
	certs := PemStringToCertificates(pem)
	if len(certs) == 0 {
		return ""
	}
	return FingerprintFromCertificate(certs[0])
}

// FingerprintFromPemString returns the sha1 fingerprint of the first parsable certificate in a pem string
func FingerprintFromPemString(pem string) string {
	certs := PemStringToCertificates(pem)
	if len(certs) == 0 {
		return ""
	}
	return FingerprintFromCertificate(certs[0])
}

// FingerprintFromPemBytes returns the sha1 fingerprint of the first parsable certificate in a pem string
func FingerprintFromPemBytes(pem []byte) string {
	certs := PemBytesToCertificates(pem)
	if len(certs) == 0 {
		return ""
	}
	return FingerprintFromCertificate(certs[0])
}

// FingerprintFromX509 returns the sha1 fingerprint of the supplied certificate
// Deprecated: Use FingerprintFromCertificate
func FingerprintFromX509(cert *x509.Certificate) string {
	return FingerprintFromCertificate(cert)
}

// FingerprintFromCertificate returns the sha1 fingerprint of the supplied certificate
func FingerprintFromCertificate(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}
	return fmt.Sprintf("%x", sha1.Sum(cert.Raw))
}
