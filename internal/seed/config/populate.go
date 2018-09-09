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
	"strconv"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/magiconair/properties"
	"github.com/pelletier/go-toml"
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

// Function to import the older, legacy configuration for services not yet converted to V2.
// Eventually, this function will be retired. Newer methodology is below.
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

// Eventually, once all services are converted to utilize ConfigV2 types, this function will replace the one above.
func ImportV2Configuration(root string, profile string) error {
	dirs := listV2Directories()
	absRoot, err := determineAbsRoot(root)
	if err != nil {
		return err
	}

	// For every application directory...
	for _, d := range dirs {
		LoggingClient.Debug(fmt.Sprintf("importing: %s/%s", absRoot, d))
		if err != nil {
			return err
		}
		// Find the resource (res) directory...
		res := fmt.Sprintf("%s/%s/res", absRoot, d)

		// Append profile to the path if specified...
		if len(profile) > 0 {
			res += "/" + profile
		}

		//Append configuration file name
		path := res + "/" + internal.ConfigFileDefaultProfile

		if !isTomlExtension(path) {
			return errors.New("unsupported file extension: " + internal.ConfigFileDefaultProfile)
		}

		LoggingClient.Debug("reading toml " + path)
		// load the ToML file
		config, _ := toml.LoadFile(path)

		// Fetch the map[string]interface{}
		m := config.ToMap()

		// traverse the map and put into KV[]
		kvs, err := traverse("", m)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("There was an error: %v", err))
		}
		for _, kv := range kvs {
			LoggingClient.Debug(fmt.Sprintf("v2 consul wrote key %s with value %s", kv.Key, kv.Value))
		}

		// Put config properties to Consul K/V store.
		prefix := internal.ConfigV2Stem + internal.ServiceKeyPrefix + d + "/"

		// Put config properties to Consul K/V store.
		for _, v := range kvs {
			p := &consulapi.KVPair{Key: prefix + v.Key, Value: []byte(v.Value)}
			if _, err := (*consulapi.KV).Put(Registry.KV(), p, nil); err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

func listDirectories() [6]string {
	var names = [6]string{internal.CoreDataServiceKey, internal.CoreMetaDataServiceKey,
		internal.ExportClientServiceKey, internal.ExportDistroServiceKey, internal.SupportLoggingServiceKey,
		internal.SupportNotificationsServiceKey}

	for i, name := range names {
		names[i] = strings.Replace(name, internal.ServiceKeyPrefix, "", 1)
	}

	return names
}

// As services are converted to utilize V2 types, add them to this list and remove from the one above.
func listV2Directories() [1]string {
	var names = [1]string{internal.CoreCommandServiceKey}

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

// This works for legacy configuration because our TOML is simply key/value.
// Will not work once we go hierarchical
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

// Key/Value pair for parsing
type pair struct {
	Key   string
	Value string
}

// Traverse or walk hierarchical configuration in preparation for loading into Consul
func traverse(path string, j interface{}) ([]*pair, error) {
	kvs := make([]*pair, 0)

	pathPre := ""
	if path != "" {
		pathPre = path + "/"
	}

	switch j.(type) {
	case []interface{}:
		for sk, sv := range j.([]interface{}) {
			skvs, err := traverse(pathPre+strconv.Itoa(sk), sv)
			if err != nil {
				return nil, err
			}
			kvs = append(kvs, skvs...)
		}
	case map[string]interface{}:
		for sk, sv := range j.(map[string]interface{}) {
			skvs, err := traverse(pathPre+sk, sv)
			if err != nil {
				return nil, err
			}
			kvs = append(kvs, skvs...)
		}
	case int:
		kvs = append(kvs, &pair{Key: path, Value: strconv.Itoa(j.(int))})
	case int64:

		var y int = int(j.(int64))

		kvs = append(kvs, &pair{Key: path, Value: strconv.Itoa(y)})
	case float64:
		kvs = append(kvs, &pair{Key: path, Value: strconv.FormatFloat(j.(float64), 'f', -1, 64)})
	case bool:
		kvs = append(kvs, &pair{Key: path, Value: strconv.FormatBool(j.(bool))})
	case nil:
		kvs = append(kvs, &pair{Key: path, Value: ""})
	default:
		kvs = append(kvs, &pair{Key: path, Value: j.(string)})
	}

	return kvs, nil
}
