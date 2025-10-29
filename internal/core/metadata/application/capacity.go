//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

func checkCapacityWithNewDevice(d models.Device, dic *di.Container) errors.EdgeX {
	config := container.ConfigurationFrom(dic.Get)
	dbClient := container.DBClientFrom(dic.Get)
	lock := container.CapacityCheckLockFrom(dic.Get)
	lock.Lock()
	defer lock.Unlock()

	if config.Writable.MaxDevices > 0 {
		deviceCount, err := dbClient.DeviceCountByLabels(nil)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), "query device count failed", err)
		}
		if deviceCount+1 > int64(config.Writable.MaxDevices) {
			return errors.NewCommonEdgeX(
				errors.KindContractInvalid,
				fmt.Sprintf("the existing total number of device is '%d', add device '%s' will exceed the maximum limitation '%d'", deviceCount, d.Name, config.Writable.MaxDevices), nil)
		}
	}
	if config.Writable.MaxResources > 0 {
		totalInUseResourceCount, err := dbClient.InUseResourceCount()
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), "query in use resource count failed", err)
		}
		newResourceCount, err := resourceCountByProfile(d.ProfileName, dic)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), "get resource count failed", err)
		}
		if totalInUseResourceCount+newResourceCount > int64(config.Writable.MaxResources) {
			return errors.NewCommonEdgeX(
				errors.KindContractInvalid,
				fmt.Sprintf("'%d' resources is in use, increase '%d' resources will exceed the maximum limitation '%d'", totalInUseResourceCount, newResourceCount, config.Writable.MaxResources), nil)
		}
	}
	return nil
}

func checkResourceCapacityByExistingAndNewProfile(oldProfileName, newProfileName string, dic *di.Container) errors.EdgeX {
	config := container.ConfigurationFrom(dic.Get)
	dbClient := container.DBClientFrom(dic.Get)
	lock := container.CapacityCheckLockFrom(dic.Get)
	lock.Lock()
	defer lock.Unlock()

	if config.Writable.MaxResources > 0 && oldProfileName != newProfileName {
		oldProfileResourceCount, err := resourceCountByProfile(oldProfileName, dic)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), "get resource count failed", err)
		}
		newProfileResourceCount, err := resourceCountByProfile(newProfileName, dic)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), "get resource count failed", err)
		}

		totalInUseResourceCount, err := dbClient.InUseResourceCount()
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), "query in use resource count failed", err)
		}
		count := totalInUseResourceCount - oldProfileResourceCount + newProfileResourceCount
		if count > int64(config.Writable.MaxResources) {
			return errors.NewCommonEdgeX(
				errors.KindContractInvalid,
				fmt.Sprintf(
					"'%d' resources is in use, change from profile '%s' (%d resource count) to the profile '%s' (%d resource count) will exceed the maximum limitation '%d'",
					totalInUseResourceCount, oldProfileName, oldProfileResourceCount, newProfileName, newProfileResourceCount, config.Writable.MaxResources), nil)
		}
	}
	return nil
}

func checkResourceCapacityByUpdateProfile(profile models.DeviceProfile, dic *di.Container) errors.EdgeX {
	config := container.ConfigurationFrom(dic.Get)
	dbClient := container.DBClientFrom(dic.Get)
	lock := container.CapacityCheckLockFrom(dic.Get)
	lock.Lock()
	defer lock.Unlock()

	isInUse, err := isProfileInUse(profile.Name, dic)
	if err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "check profile in use failed", err)
	}
	if isInUse {
		totalInUseResourceCount, err := dbClient.InUseResourceCount()
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), "query in use resource count failed", err)
		}
		existingProfileResourceCount, err := resourceCountByProfile(profile.Name, dic)
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), "query existing profile resource count failed", err)
		}
		newProfileResourceCount := int64(len(profile.DeviceResources))
		count := totalInUseResourceCount - existingProfileResourceCount + newProfileResourceCount
		if count > int64(config.Writable.MaxResources) {
			return errors.NewCommonEdgeX(
				errors.KindContractInvalid,
				fmt.Sprintf(
					"'%d' resources is in use, update the profile from %d resource count to %d resource count will exceed the maximum limitation '%d'",
					totalInUseResourceCount, existingProfileResourceCount, newProfileResourceCount, config.Writable.MaxResources), nil)
		}
	}
	return nil
}

func checkResourceCapacityByNewResource(profileName string, resource models.DeviceResource, dic *di.Container) errors.EdgeX {
	config := container.ConfigurationFrom(dic.Get)
	dbClient := container.DBClientFrom(dic.Get)
	lock := container.CapacityCheckLockFrom(dic.Get)
	lock.Lock()
	defer lock.Unlock()

	isInUse, err := isProfileInUse(profileName, dic)
	if err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "check profile in use failed", err)
	}
	if isInUse {
		totalInUseResourceCount, err := dbClient.InUseResourceCount()
		if err != nil {
			return errors.NewCommonEdgeX(errors.Kind(err), "query in use resource count failed", err)
		}
		if totalInUseResourceCount+1 > int64(config.Writable.MaxResources) {
			return errors.NewCommonEdgeX(
				errors.KindContractInvalid,
				fmt.Sprintf(
					"'%d' resources is in use, add '%s' resource will exceed the maximum limitation '%d'",
					totalInUseResourceCount, resource.Name, config.Writable.MaxResources), nil)
		}
	}
	return nil
}

func resourceCountByProfile(profileName string, dic *di.Container) (int64, errors.EdgeX) {
	if profileName == "" {
		return 0, nil
	}
	dbClient := container.DBClientFrom(dic.Get)
	profile, err := dbClient.DeviceProfileByName(profileName)
	if err != nil {
		return 0, errors.NewCommonEdgeX(errors.Kind(err), "count resource number failed", err)
	}
	return int64(len(profile.DeviceResources)), nil
}
