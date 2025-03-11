// +build windows

/*
	Copyright 2019 NetFoundry Inc.

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

package posture

import (
	"crypto/sha1"
	"debug/pe"
	"fmt"
	"go.mozilla.org/pkcs7"
	"os"
	"strings"
)

func isProcessPath(expectedPath, processPath string) bool {
	return strings.Contains(strings.ToLower(expectedPath), strings.ToLower(processPath))
}

func getSignerFingerprints(filePath string) ([]string, error) {
	peFile, err := pe.Open(filePath)

	if err != nil {
		return nil, err
	}
	defer peFile.Close()

	var virtualAddress uint32
	var size uint32
	switch t := peFile.OptionalHeader.(type) {
	case *pe.OptionalHeader32:
		virtualAddress = t.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_SECURITY].VirtualAddress
		size = t.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_SECURITY].Size
	case *pe.OptionalHeader64:
		virtualAddress = t.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_SECURITY].VirtualAddress
		size = t.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_SECURITY].Size
	}

	if virtualAddress <= 0 || size <= 0 {
		return nil, nil
	}

	binaryFile, err := os.Open(filePath)

	if err != nil {
		return nil, fmt.Errorf("failed to open process file: %v", err)
	}
	defer binaryFile.Close()

	pkcs7Signature := make([]byte, int64(size))
	_, _ = binaryFile.ReadAt(pkcs7Signature, int64(virtualAddress+8))

	pkcs7Obj, err := pkcs7.Parse(pkcs7Signature)

	if err != nil {
		return nil, fmt.Errorf("failed to parse pkcs7 block: %v", err)
	}
	var signerFingerprints []string
	for _, cert := range pkcs7Obj.Certificates {
		signerFingerprint := fmt.Sprintf("%x", sha1.Sum(cert.Raw))
		signerFingerprints = append(signerFingerprints, signerFingerprint)
	}

	return signerFingerprints, nil
}
