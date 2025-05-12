// Copyright (C) 2025 IOTech Ltd

package utils

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// ValidateEvent checks if the incoming event's profileName, deviceName and sourceName match the message topic where the event received from
func ValidateEvent(messageTopic string, e dtos.Event) errors.EdgeX {
	// Parse messageTopic by the pattern `edgex/events/device/<device-service-name>/<device-profile-name>/<device-name>/<source-name>`
	fields := strings.Split(messageTopic, "/")

	// assumes a non-empty base topic with events/device/<device-service-name>/<device-profile-name>/<device-name>/<source-name>
	if len(fields) < 6 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("invalid message topic %s", messageTopic), nil)
	}

	len := len(fields)
	profileName, err := url.PathUnescape(fields[len-3])
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	deviceName, err := url.PathUnescape(fields[len-2])
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	sourceName, err := url.PathUnescape(fields[len-1])
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	// Check whether the event fields match the message topic
	if e.ProfileName != profileName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's profileName %s mismatches with the name %s received in topic", e.ProfileName, profileName), nil)
	}
	if e.DeviceName != deviceName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's deviceName %s mismatches with the name %s received in topic", e.DeviceName, deviceName), nil)
	}
	if e.SourceName != sourceName {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("event's sourceName %s mismatches with the name %s received in topic", e.SourceName, sourceName), nil)
	}
	return nil
}
