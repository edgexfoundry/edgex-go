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
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/magiconair/properties"
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

func ImportConfiguration(root string) error {
	dirs := listDirectories()
	absRoot, err := determineAbsRoot(root)
	if err != nil {
		return err
	}
	for _, d := range dirs {
		LoggingClient.Debug(fmt.Sprintf("importing: %s/%s", absRoot, d))
		if err != nil {
			return err
		}
		res := fmt.Sprintf("%s/%s/res", absRoot, d)
		err = filepath.Walk(res, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Only import TOML files. Ignore others.
			if info.IsDir() || !isTomlExtension(info.Name()) {
				return nil
			}

			LoggingClient.Debug("reading toml " + path)
			props, err := readTomlFile(path)
			if err != nil {
				return err
			}

			appKey, err := buildAppKey(d, info.Name())
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
	}
	return nil
}

func listDirectories() [7]string {
	var names = [7]string{internal.CoreCommandServiceKey, internal.CoreDataServiceKey, internal.CoreMetaDataServiceKey,
		internal.ExportClientServiceKey, internal.ExportDistroServiceKey, internal.SupportLoggingServiceKey,
		internal.SupportNotificationsServiceKey}

	for i, name := range names {
		names[i] = strings.Replace(name, internal.ServiceKeyPrefix, "", 1)
	}

	return names
}

func determineAbsRoot(root string) (string, error) {
	// This should only be necessary when running the application locally as a developer
	// so some knowledge of the tree layout is granted.
	if strings.Contains(root, "../cmd") {
		exec, err := os.Executable()
		if err != nil {
			return "", err
		}
		execDir := filepath.Dir(exec)
		LoggingClient.Debug("executing dir " + execDir)
		ixLast := strings.LastIndex(execDir, "/")
		absRoot := execDir[0:(ixLast)] //Seems like there should be some other way
		LoggingClient.Debug("absolute path " + absRoot)
		return absRoot, nil
	}
	return root, nil
}

func buildAppKey(app string, fileName string) (string, error) {
	var err error = nil
	var key string = internal.ServiceKeyPrefix
	switch fileName {
	case internal.ConfigFileDefaultProfile:
		key += app + ";go"
		break
	case internal.ConfigFileDockerProfile:
		key += app + ";docker"
		break
	default:
		err = errors.New("unrecognized name: " + fileName)
		key = ""
		break
	}
	return key, err
}

func isAcceptablePropertyExtensions(file string) bool {
	for _, v := range Configuration.AcceptablePropertyExtensions {
		if v == filepath.Ext(file) {
			return true
		}
	}
	return false
}

func isTomlExtension(file string) bool {
	for _, v := range Configuration.TomlExtensions {
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

//This works for now because our TOML is simply key/value.
//Will not work once we go hierarchical
func readTomlFile(filePath string) (map[string]string, error) {
	configProps := map[string]string{}

	file, err := os.Open(filePath)
	if err != nil {
		return configProps, fmt.Errorf("could not load configuration file (%s): %v", filePath, err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			tokens := strings.Split(scanner.Text(), "=")
			configProps[strings.Trim(tokens[0], " '")] = strings.Trim(tokens[1], " '")
		}
	}
	return configProps, nil
}
