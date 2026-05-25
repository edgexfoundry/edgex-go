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

package certtools

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var CURVES = map[string]elliptic.Curve{}

var curves = []elliptic.Curve{
	elliptic.P224(),
	elliptic.P256(),
	elliptic.P384(),
	elliptic.P521(),
}

func init() {
	for _, c := range curves {
		CURVES[c.Params().Name] = c
	}
}

// GetKey will attempt to load an engine key from `eng` if provided. If `eng` is `nil`, `file` and `newkey` will be
// evaluated. `file` will be loaded; if the file does not exist a new key according to `newkey` will be created.
// If it does exist, its key type (RSA bit size, EC curve) will be compared to `newkey`. If the desired type does not
// match the loaded type an error will be returned.
func GetKey(eng *url.URL, file, newkey string) (crypto.PrivateKey, error) {
	if eng != nil {
		var engine = eng.Scheme
		return LoadEngineKey(engine, eng)
	}

	var existingKey crypto.PrivateKey
	if file != "" {
		if pemBytes, err := os.ReadFile(file); err == nil {
			existingKey, err = LoadPrivateKey(pemBytes)

			if err != nil {
				return nil, fmt.Errorf("detected existing key [%s] and failed to load it: %w", file, err)
			}

			//no type specified, return it
			if newkey == "" {
				return existingKey, nil
			}

			//verify that it matches what we want, otherwise error
			return verifyExistingKey(file, existingKey, newkey)
		} else if !os.IsNotExist(err) {
			// if the error is anything but "does not exist" we do not know what to do.
			return nil, fmt.Errorf("could not read file [%s]: %w", file, err)
		}

		if newkey == "" {
			return nil, fmt.Errorf("no file found at [%s] but a new key spec was not provided", file)
		}

		//no file exists, but we have a key spec, will generate below
	}

	if newkey != "" {
		key, err := generateKey(newkey)
		if err != nil {
			return nil, err
		}

		if err := SavePrivateKey(key, file); err != nil {
			return nil, err
		}

		return key, nil
	}

	if file != "" {
		if pemBytes, err := os.ReadFile(file); err != nil {
			return nil, err
		} else {
			return LoadPrivateKey(pemBytes)
		}
	}

	return nil, fmt.Errorf("no key mechanism specified")
}

func verifyExistingKey(file string, existingKey crypto.PrivateKey, newkey string) (crypto.PrivateKey, error) {
	//desired type specified, verify it
	specs := strings.Split(newkey, ":")

	if len(specs) != 2 {
		return nil, fmt.Errorf("invalid new key spec, got: %s, need format of: <[EC|RSA]]>:<[BitSize|Curve]>", newkey)
	}

	switch t := existingKey.(type) {
	case *ecdsa.PrivateKey:
		if specs[0] != "ec" {
			return nil, fmt.Errorf("detected existing key [%s] but was of the wrong type, expected an EC key", file)
		}

		if t.Curve != CURVES[specs[1]] {
			return nil, fmt.Errorf("detected existing key [%s] but was of the wrong curve type: %s, expected: %s", file, t.Curve.Params().Name, specs[1])
		}

		return existingKey, nil

	case *rsa.PrivateKey:
		if specs[0] != "rsa" {
			return nil, fmt.Errorf("detected existing key [%s] but was of the wrong type, expected an RSA key", file)
		}
		bitSize, err := strconv.Atoi(specs[1])

		if err != nil {
			return nil, fmt.Errorf("error parsing RSA bit size from new key spec, got: %s, need format of: <[EC|RSA]]>:<[BitSize|Curve]>", newkey)
		}
		if (t.PublicKey.Size() * 8) != bitSize {
			return nil, fmt.Errorf("detected existing key [%s] but was of wrong bit size: %d, expected: %d", file, t.PublicKey.Size(), bitSize)
		}

		return existingKey, nil
	default:
		return nil, fmt.Errorf("detected existing key [%s] which is an unsupported type: %T", file, existingKey)
	}
}

func SavePrivateKey(key crypto.PrivateKey, file string) error {
	var der []byte
	var t string
	if rsaK, ok := key.(*rsa.PrivateKey); ok {
		t = "RSA PRIVATE KEY"
		der = x509.MarshalPKCS1PrivateKey(rsaK)
	} else if ecK, ok := key.(*ecdsa.PrivateKey); ok {
		t = "EC PRIVATE KEY"
		der, _ = x509.MarshalECPrivateKey(ecK)
	} else {
		return fmt.Errorf("Unsupported key type")
	}

	keyPem := &pem.Block{Type: t, Bytes: der}

	return os.WriteFile(file, pem.EncodeToMemory(keyPem), 0600)
}

func LoadPrivateKey(pemBytes []byte) (crypto.PrivateKey, error) {

	var keyBlock *pem.Block
	for len(pemBytes) > 0 {
		keyBlock, pemBytes = pem.Decode(pemBytes)
		switch keyBlock.Type {
		case "EC PRIVATE KEY":
			return x509.ParseECPrivateKey(keyBlock.Bytes)
		case "RSA PRIVATE KEY":
			return x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		case "PRIVATE KEY":
			return x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		}
	}

	return nil, fmt.Errorf("no key found")
}

func SupportedCurves() []string {
	names := make([]string, 0, len(curves))
	for _, c := range curves {
		names = append(names, c.Params().Name)
	}
	return names
}

func generateKey(spec string) (crypto.PrivateKey, error) {
	specs := strings.Split(spec, ":")

	switch specs[0] {
	case "rsa":
		if bits, err := strconv.Atoi(specs[1]); err != nil {
			return nil, err
		} else {
			return rsa.GenerateKey(rand.Reader, bits)
		}
	case "ec":
		if c, ok := CURVES[specs[1]]; !ok {
			return nil, fmt.Errorf("ECurve '%s' not found", specs[1])
		} else {
			return ecdsa.GenerateKey(c, rand.Reader)
		}
	default:
		return nil, fmt.Errorf("unsupported key spec '%s'", specs[0])
	}
}
