//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package requests

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	edgexErrors "github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

// DeviceProfileTagsRequest defines the Request Content for PATCH UpdateDeviceProfileTags DTO.
type DeviceProfileTagsRequest struct {
	dtoCommon.BaseRequest        `json:",inline"`
	dtos.UpdateDeviceProfileTags `json:",inline"`
}

// Validate satisfies the Validator interface
func (dt *DeviceProfileTagsRequest) Validate() error {
	err := common.Validate(dt)
	if err != nil {
		// The UpdateDeviceProfileTags is the internal struct in Golang programming, not in the Profile model,
		// so it should be hidden from the error messages.
		err = errors.New(strings.ReplaceAll(err.Error(), ".UpdateDeviceProfileTags", ""))
		return edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid, "", err)
	}
	return nil
}

// UnmarshalJSON implements the Unmarshaler interface for the DeviceProfileTagsRequest type
func (dt *DeviceProfileTagsRequest) UnmarshalJSON(b []byte) error {
	var alias struct {
		dtoCommon.BaseRequest
		dtos.UpdateDeviceProfileTags
	}
	if err := json.Unmarshal(b, &alias); err != nil {
		return edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid, "Failed to unmarshal request body as JSON.", err)
	}

	*dt = DeviceProfileTagsRequest(alias)

	// validate DeviceProfileTagsRequest DTO
	if err := dt.Validate(); err != nil {
		return err
	}
	return nil
}

// ReplaceDeviceProfileModelTagsWithDTO replace existing DeviceProfile's device resources/commands tags fields with DTO patch
func ReplaceDeviceProfileModelTagsWithDTO(dp *models.DeviceProfile, patch dtos.UpdateDeviceProfileTags) {
	// Convert UpdateTags list to maps for fast lookup
	resourceUpdate := convertUpdateTagsToMap(patch.DeviceResources)
	commandUpdate := convertUpdateTagsToMap(patch.DeviceCommands)

	// Update Device Resource Tags
	for i := range dp.DeviceResources {
		r := &dp.DeviceResources[i]
		if tags, ok := resourceUpdate[r.Name]; ok && tags != nil {
			r.Tags = mergeTags(r.Tags, tags)
		}
	}

	// Update Device Command Tags
	for i := range dp.DeviceCommands {
		c := &dp.DeviceCommands[i]
		if tags, ok := commandUpdate[c.Name]; ok && tags != nil {
			c.Tags = mergeTags(c.Tags, tags)
		}
	}
}

func convertUpdateTagsToMap(list []dtos.UpdateTags) map[string]map[string]any {
	tagsMap := make(map[string]map[string]any)
	for _, u := range list {
		tagsMap[u.Name] = u.Tags
	}

	return tagsMap
}

func mergeTags(dest, src map[string]any) map[string]any {
	if dest == nil {
		dest = make(map[string]any)
	}
	for key, value := range src {
		dv, destOk := dest[key].(map[string]any)
		sv, srcOk := value.(map[string]any)
		if destOk && srcOk {
			dest[key] = mergeTags(dv, sv)
			continue
		}

		dest[key] = value
	}
	return dest
}
