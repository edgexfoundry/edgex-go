/*******************************************************************************
 * Copyright 2023 Intel Corp.
 * Copyright (C) 2023 IOTech Ltd
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

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/labstack/echo/v4"
)

const PathSep = "/"

// ConvertToMap uses json to marshal and unmarshal a target type into a map
func ConvertToMap(target any, m *map[string]any) error {
	jsonBytes, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("could not marshal %T to JSON: %v", target, err)
	}
	if err = json.Unmarshal(jsonBytes, &m); err != nil {
		return fmt.Errorf("could not unmarshal JSON (from %T) into a map: %v", target, err)
	}
	return nil
}

// ConvertFromMap uses json to marshal and unmarshal a map into a target type
func ConvertFromMap(m map[string]any, target any) error {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("could not marshal map to JSON: %v", err)
	}

	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("could not unmarshal JSON to %T: %v", target, err)
	}

	return nil
}

// MergeMaps combines the src map keys and values with the dest map keys and values if the key exists
func MergeMaps(dest map[string]any, src map[string]any) {

	var exists bool

	for key, value := range src {
		_, exists = dest[key]
		if !exists {
			dest[key] = value
			continue
		}

		destVal, ok := dest[key].(map[string]any)
		if ok {
			MergeMaps(destVal, value.(map[string]any))
			continue
		}

		dest[key] = value
	}
}

func RemoveUnusedSettings(src any, baseKey string, usedSettingKeys map[string]any) (map[string]any, error) {
	srcMap := make(map[string]any)

	if err := ConvertToMap(src, &srcMap); err != nil {
		return nil, fmt.Errorf("could not create map from %T: %s", src, err.Error())
	}

	removeUnusedSettingsFromMap(srcMap, baseKey, usedSettingKeys)

	return srcMap, nil
}

// removeMapUnusedSettings iterates over a map and removes any settings not in list of valid keys
func removeUnusedSettingsFromMap(target map[string]any, baseKey string, validKeys map[string]any) {
	var removeKeys []string
	for key, value := range target {
		nextBaseKey := BuildBaseKey(baseKey, key)
		sub, ok := value.(map[string]any)
		if ok {
			removeUnusedSettingsFromMap(sub, nextBaseKey, validKeys)
			if len(sub) == 0 {
				removeKeys = append(removeKeys, key)
			}
			continue
		}
		_, exists := validKeys[nextBaseKey]
		if !exists {
			removeKeys = append(removeKeys, key)
		}
	}

	for _, key := range removeKeys {
		delete(target, key)
	}
}

// MergeValues combines src with the dest.
func MergeValues(dest any, src any) error {
	var ok bool
	var destMap, srcMap map[string]any

	destMap, ok = dest.(map[string]any)
	if !ok {
		if err := ConvertToMap(dest, &destMap); err != nil {
			return fmt.Errorf("could not create destination map from %T: %s", dest, err.Error())
		}
	}

	srcMap, ok = src.(map[string]any)
	if !ok {
		if err := ConvertToMap(src, &srcMap); err != nil {
			return fmt.Errorf("could not source create map from %T: %s", src, err.Error())
		}
	}

	MergeMaps(destMap, srcMap)

	// convert the map back to a dest
	if err := ConvertFromMap(destMap, dest); err != nil {
		return err
	}

	return nil
}

func StringSliceToMap(src []string) map[string]any {
	result := make(map[string]any)

	for _, value := range src {
		result[value] = nil
	}

	return result
}

func BuildBaseKey(keys ...string) string {
	return strings.Join(keys, PathSep)
}

// DeepCopy creates a deep copy/clone of a struct by using json to marshal the original struct, and then unmarshal it
// back into the new copy. Note that this will only copy the exported fields.
func DeepCopy(src any, dest any) error {
	jsonBytes, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("could not marshal %T to JSON: %v", src, err)
	}
	if err = json.Unmarshal(jsonBytes, &dest); err != nil {
		return fmt.Errorf("could not unmarshal JSON (from %T) into type %T: %v", src, dest, err)
	}
	return nil
}

// SendJsonResp puts together the response packet for the APIs
func SendJsonResp(
	lc logger.LoggingClient,
	writer *echo.Response,
	request *http.Request,
	response interface{},
	statusCode int) error {

	correlationID := request.Header.Get(common.CorrelationHeader)

	writer.Header().Set(common.CorrelationHeader, correlationID)
	writer.Header().Set(common.ContentType, common.ContentTypeJSON)
	writer.WriteHeader(statusCode)

	if response != nil {
		data, err := json.Marshal(response)
		if err != nil {
			lc.Error("Unable to marshal response", "error", err.Error(), common.CorrelationHeader, correlationID)
			// set Response.Committed to true in order to rewrite the status code
			writer.Committed = false
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		_, err = writer.Write(data)
		if err != nil {
			lc.Error("Unable to marshal response", "error", err.Error(), common.CorrelationHeader, correlationID)
			// set Response.Committed to true in order to rewrite the status code
			writer.Committed = false
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	return nil
}

// SendJsonErrResp puts together the error response packet for the APIs
func SendJsonErrResp(
	lc logger.LoggingClient,
	writer *echo.Response,
	request *http.Request,
	errKind errors.ErrKind,
	message string,
	err error,
	requestID string) error {

	edgeXerr := errors.NewCommonEdgeX(errKind, message, err)
	lc.Error(edgeXerr.Error())
	lc.Debug(edgeXerr.DebugMessages())
	response := commonDTO.NewBaseResponse(requestID, edgeXerr.Message(), edgeXerr.Code())
	return SendJsonResp(lc, writer, request, response, edgeXerr.Code())
}
