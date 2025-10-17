//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package constants

import (
	"regexp"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
)

// new constants relates to EdgeX Keeper service and will be added to go-mod-core-contracts in the future
const CoreKeeperServiceKey = "core-keeper"

// KeyAllowedCharsRegexString defined the characters allowed in the key name
const KeyAllowedCharsRegexString = "^[a-zA-Z0-9-_~;=./%]+$"

var (
	KeyAllowedCharsRegex = regexp.MustCompile(KeyAllowedCharsRegexString)
)

// key delimiter for edgex keeper
const KeyDelimiter = "/"

// Constants related to defined routes in the v3 service APIs
const ApiKVRoute = common.ApiBase + "/kvs/" + Key + "/{" + Key + ":.*}"
const ApiRegisterRoute = common.ApiBase + "/registry"
const ApiAllRegistrationsRoute = ApiRegisterRoute + "/" + common.All
const ApiRegistrationByServiceIdRoute = ApiRegisterRoute + "/" + ServiceId + "/{" + ServiceId + "}"

// Constants related to defined url path names and parameters in the v2 service APIs
const (
	Flatten      = "flatten"
	Key          = "key"
	KeyOnly      = "keyOnly"
	Plaintext    = "plaintext"
	PrefixMatch  = "prefixMatch"
	ServiceId    = "serviceId"
	Deregistered = "deregistered"
)
