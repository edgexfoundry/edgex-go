/*
 * ******************************************************************************
 *  Copyright 2019 Dell Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software distributed under the License
 *  is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 *  or implied. See the License for the specific language governing permissions and limitations under
 *  the License.
 *  ******************************************************************************
 */

package value_descriptor

import (
	"github.com/edgexfoundry/edgex-go/internal/core/data/errors"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/config"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

// GetValueDescriptorsExecutor retrieves one or more value descriptors.
type GetValueDescriptorsExecutor interface {
	Execute() ([]contract.ValueDescriptor, error)
}

// getValueDescriptorsByNames encapsulates data needed to retrieve value descriptors by one or more names.
type getValueDescriptorsByNames struct {
	loader Loader
	names  []string
	logger logger.LoggingClient
	config bootstrapConfig.ServiceInfo
}

// Execute retrieves value descriptors by one or more names.
func (g getValueDescriptorsByNames) Execute() ([]contract.ValueDescriptor, error) {
	vds, err := g.loader.ValueDescriptorsByName(g.names)
	if err != nil {
		return nil, err
	}

	if len(vds) > g.config.MaxResultCount {
		return nil, errors.NewErrLimitExceeded(len(vds))
	}

	return vds, nil
}

// NewGetValueDescriptorsNameExecutor creates a GetValueDescriptorsExecutor which will get value descriptors matching the provided names.
func NewGetValueDescriptorsNameExecutor(names []string, loader Loader, logger logger.LoggingClient, config bootstrapConfig.ServiceInfo) GetValueDescriptorsExecutor {
	return getValueDescriptorsByNames{
		names:  names,
		logger: logger,
		loader: loader,
		config: config,
	}
}

// getAllValueDescriptors encapsulates the data needed to retrieve all value descriptors.
type getAllValueDescriptors struct {
	loader Loader
	logger logger.LoggingClient
	config bootstrapConfig.ServiceInfo
}

// Execute retrieves value descriptors by the provided names.
func (g getAllValueDescriptors) Execute() ([]contract.ValueDescriptor, error) {
	vds, err := g.loader.ValueDescriptors()
	if err != nil {
		return nil, err
	}

	if len(vds) > g.config.MaxResultCount {
		return nil, errors.NewErrLimitExceeded(len(vds))
	}

	return vds, nil
}

// NewGetValueDescriptorsExecutor creates a GetValueDescriptorsExecutor which will get all value descriptors.
func NewGetValueDescriptorsExecutor(loader Loader, logger logger.LoggingClient, config bootstrapConfig.ServiceInfo) GetValueDescriptorsExecutor {
	return getAllValueDescriptors{
		logger: logger,
		loader: loader,
		config: config,
	}
}
