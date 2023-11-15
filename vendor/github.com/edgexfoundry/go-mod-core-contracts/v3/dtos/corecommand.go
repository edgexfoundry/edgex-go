//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

type DeviceCoreCommand struct {
	DeviceName   string        `json:"deviceName" validate:"required,edgex-dto-none-empty-string"`
	ProfileName  string        `json:"profileName" validate:"required,edgex-dto-none-empty-string"`
	CoreCommands []CoreCommand `json:"coreCommands,omitempty" validate:"dive"`
}

type CoreCommand struct {
	Name       string                 `json:"name" validate:"required,edgex-dto-none-empty-string"`
	Get        bool                   `json:"get,omitempty" validate:"required_without=Set"`
	Set        bool                   `json:"set,omitempty" validate:"required_without=Get"`
	Path       string                 `json:"path,omitempty"`
	Url        string                 `json:"url,omitempty"`
	Parameters []CoreCommandParameter `json:"parameters,omitempty"`
}

type CoreCommandParameter struct {
	ResourceName string `json:"resourceName"`
	ValueType    string `json:"valueType"`
}
