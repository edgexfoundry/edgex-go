//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

// AddProvisionWatcher adds a new provision watcher
func (c *Client) AddProvisionWatcher(pw model.ProvisionWatcher) (provisionwatcher model.ProvisionWatcher, err errors.EdgeX) {
	return provisionwatcher, nil
}

// ProvisionWatcherById gets a provision watcher by id
func (c *Client) ProvisionWatcherById(id string) (provisionwatcher model.ProvisionWatcher, err errors.EdgeX) {
	return provisionwatcher, nil
}

// ProvisionWatcherByName gets a provision watcher by name
func (c *Client) ProvisionWatcherByName(name string) (provisionwatcher model.ProvisionWatcher, err errors.EdgeX) {
	return provisionwatcher, nil
}

// ProvisionWatchersByServiceName query provision watchers by offset, limit and service name
func (c *Client) ProvisionWatchersByServiceName(offset int, limit int, name string) (pws []model.ProvisionWatcher, err errors.EdgeX) {
	return pws, nil
}

// ProvisionWatchersByProfileName query provision watchers by offset, limit and profile name
func (c *Client) ProvisionWatchersByProfileName(offset int, limit int, name string) (pws []model.ProvisionWatcher, err errors.EdgeX) {
	return pws, nil
}

// AllProvisionWatchers query provision watchers with offset, limit and labels
func (c *Client) AllProvisionWatchers(offset int, limit int, labels []string) (pws []model.ProvisionWatcher, err errors.EdgeX) {
	return pws, nil
}

// DeleteProvisionWatcherByName deletes a provision watcher by name
func (c *Client) DeleteProvisionWatcherByName(name string) errors.EdgeX {
	return nil
}

// Update a provision watcher
func (c *Client) UpdateProvisionWatcher(pw model.ProvisionWatcher) errors.EdgeX {
	return nil
}

// ProvisionWatcherCountByLabels returns the total count of Provision Watchers with labels specified.  If no label is specified, the total count of all provision watchers will be returned.
func (c *Client) ProvisionWatcherCountByLabels(labels []string) (count uint32, err errors.EdgeX) {
	return count, nil
}

// ProvisionWatcherCountByServiceName returns the count of Provision Watcher associated with specified service
func (c *Client) ProvisionWatcherCountByServiceName(name string) (count uint32, err errors.EdgeX) {
	return count, nil
}

// ProvisionWatcherCountByProfileName returns the count of Provision Watcher associated with specified profile
func (c *Client) ProvisionWatcherCountByProfileName(name string) (count uint32, err errors.EdgeX) {
	return count, nil
}
