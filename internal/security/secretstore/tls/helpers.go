//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"fmt"
	"os"
	"os/exec"
)

func checkIfFileExists(fileName string) bool {
	fileInfo, statErr := os.Stat(fileName)
	if os.IsNotExist(statErr) {
		return false
	}
	return !fileInfo.IsDir()
}

func runOpensslCmd(args []string) error {
	cmd := exec.Command("openssl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func createOpensslConf(opensslConfFilePath string, hostName string) error {
	opensslConf, err := os.Create(opensslConfFilePath)
	if err != nil {
		return fmt.Errorf("fail to create openssl config file, err: %v", err)
	}
	defer opensslConf.Close()
	conf := fmt.Sprintf("subjectAltName=DNS:%s", hostName)
	_, err = opensslConf.Write([]byte(conf))
	if err != nil {
		return fmt.Errorf("fail to write openssl config, err: %v", err)
	}
	return nil
}

func readSecretFromFile(namePathPair map[string]string) (map[string]string, error) {
	secrets := map[string]string{}
	for name, path := range namePathPair {
		contents, err := os.ReadFile(path)
		if err == nil {
			secrets[name] = string(contents)
		} else {
			return nil, err
		}
	}
	return secrets, nil
}
