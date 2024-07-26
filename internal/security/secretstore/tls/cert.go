//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"fmt"
	"os"
	"path/filepath"
)

func createCA(commonName, certPath, keyPath string) error {
	caSubject := fmt.Sprintf("/C=UK/O=IOTech/CN=%s", commonName)
	return runOpensslCmd([]string{"req", "-newkey", "rsa:2048", "-new", "-nodes", "-x509", "-days", "36500", "-keyout", keyPath,
		"-subj", caSubject, "-out", certPath})
}

func createServerCerts(hostname, outputDir, keyFileName, certFileName string) (caCertFilePath, caKeyFilePath string, err error) {
	opensslConfFilePath := filepath.Join(outputDir, opensslConfFileName)
	caKeyFilePath = filepath.Join(outputDir, caKeyFileName)
	caCertFilePath = filepath.Join(outputDir, caCertFileName)
	keyFilePath := filepath.Join(outputDir, keyFileName)
	csrFilePath := filepath.Join(outputDir, hostname+".csr")
	certFilePath := filepath.Join(outputDir, certFileName)

	// Check whether the certificates exist
	if checkIfFileExists(caCertFilePath) &&
		checkIfFileExists(caKeyFilePath) &&
		checkIfFileExists(certFilePath) &&
		checkIfFileExists(keyFilePath) {
		return
	}

	// Create the output folder with 0750 permission(user/owner can read, write or execute.) for server certificates
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		err = fmt.Errorf("failed to create folder for server certificates, err: %v", err)
		return
	}

	// Create CA certificate and key
	err = createCA(hostname, caCertFilePath, caKeyFilePath)
	if err != nil {
		err = fmt.Errorf("failed to generate the TLS CA key and certificate, err: %v", err)
		return
	}

	// Create a certificate signing request (CSR) for the server
	subj := fmt.Sprintf("/CN=%s", hostname)
	err = runOpensslCmd([]string{"req", "-newkey", "rsa:2048", "-nodes", "-keyout", keyFilePath,
		"-subj", subj, "-out", csrFilePath})
	if err != nil {
		err = fmt.Errorf("failed to generate the TLS certificate signing request, err: %v", err)
		return
	}

	// Sign the server CSR with the CA key and cert
	err = createOpensslConf(opensslConfFilePath, hostname)
	if err != nil {
		err = fmt.Errorf("failed to create the openssl config, err: %v", err)
		return
	}
	err = runOpensslCmd([]string{"x509", "-req", "-extfile", opensslConfFilePath, "-days", "36500",
		"-in", csrFilePath, "-CA", caCertFilePath, "-CAkey", caKeyFilePath, "-CAcreateserial", "-out", certFilePath})
	if err != nil {
		err = fmt.Errorf("failed to sign the server CSR, err: %v", err)
		return
	}

	if err = os.Chmod(caCertFilePath, 0755); err != nil {
		err = fmt.Errorf("failed to change ca cert file permission, err: %v", err)
		return
	}
	if err = os.Chmod(certFilePath, 0755); err != nil {
		err = fmt.Errorf("failed to change server cert file permission, err: %v", err)
		return
	}
	if err = os.Chmod(keyFilePath, 0755); err != nil {
		err = fmt.Errorf("failed to change server key file permission, err: %v", err)
		return
	}
	return
}
