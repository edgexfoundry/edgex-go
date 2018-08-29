/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
package config

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/magiconair/properties"
	"os"
	"path/filepath"
	"strings"
)

func ImportProperties(root string) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories & unacceptable property extension
		if info.IsDir() || !isAcceptablePropertyExtensions(info.Name()) {
			return nil
		}

		dir, file := filepath.Split(path)
		appKey := parseDirectoryName(dir)
		LoggingClient.Info(fmt.Sprintf("dir: %s file: %s", appKey, file))
		props, err := readPropertiesFile(path)
		if err != nil {
			return err
		}

		// Put config properties to Consul K/V store.
		prefix := Configuration.GlobalPrefix + "/" + appKey + "/"
		for k := range props {
			p := &consulapi.KVPair{Key: prefix + k, Value: []byte(props[k])}
			if _, err := (*consulapi.KV).Put(Registry.KV(), p, nil); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func isAcceptablePropertyExtensions(file string) bool {
	for _, v := range Configuration.AcceptablePropertyExtensions {
		if v == filepath.Ext(file) {
			return true
		}
	}
	return false
}

func parseDirectoryName(path string) string {
	path = strings.TrimRight(path, "/")
	tokens := strings.Split(path, "/")
	return tokens[len(tokens)-1]
}

// Parse a properties file to a map.
func readPropertiesFile(filePath string) (map[string]string, error) {
	props, err := properties.LoadFile(filePath, properties.UTF8)
	if err != nil {
		return nil, err
	}

	return props.Map(), nil
}
