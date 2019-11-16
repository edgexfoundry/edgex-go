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

package directory

import (
	"fmt"
	"os"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

type Handler interface {
	Create(path string) error
	Verify(path string) error
}

type handler struct {
	logger logger.LoggingClient
}

func NewHandler(logger logger.LoggingClient) Handler {
	return handler{logger: logger}
}

func (h handler) Create(path string) error {
	// Remove eventual previous PKI setup directory
	// Create a new empty PKI setup directory
	h.logger.Debug("New CA creation requested by configuration")
	h.logger.Debug("Cleaning up CA PKI setup directory")

	err := os.RemoveAll(path) // Remove pkiCaDir
	if err != nil {
		return fmt.Errorf("Attempted removal of existing CA PKI config directory: %s (%s)", path, err)
	}

	h.logger.Debug("Creating CA PKI setup directory: %s", path)
	err = os.MkdirAll(path, 0750) // Create pkiCaDir
	if err != nil {
		return fmt.Errorf("Failed to create the CA PKI configuration directory: %s (%s)", path, err)
	}
	return nil
}

func (h handler) Verify(path string) error {
	h.logger.Debug("No new CA creation requested by configuration")

	// Is the CA there? (if nil then OK... but could be something else than a directory)
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("CA PKI setup directory does not exist: %s", path)
		}
		return fmt.Errorf("CA PKI setup directory cannot be reached: %s (%s)", path, err)
	}
	if stat.IsDir() {
		h.logger.Debug(fmt.Sprintf("Existing CA PKI setup directory: %s", path))
	} else {
		return fmt.Errorf("Existing CA PKI setup directory is not a directory: %s", path)
	}
	return nil
}
