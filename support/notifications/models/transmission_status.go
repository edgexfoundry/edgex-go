/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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
 *
 *******************************************************************************/

package models

import (
	"encoding/json"
	"fmt"
)

type TransmissionStatus string

const (
	Failed       = "FAILED"
	Sent         = "SENT"
	Acknowledged = "ACKNOWLEDGED"
	Trxescalated = "TRXESCALATED"
)

/*
 *  Unmarshal the enum type
 */
func (as *TransmissionStatus) UnmarshalJSON(data []byte) error {
	// Extract the string from data.
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("TransmissionStatus should be a string, got %s", data)
	}

	got, err := map[string]TransmissionStatus{"FAILED": Failed, "SENT": Sent, "ACKNOWLEDGED": Acknowledged, "TRXESCALATED": Trxescalated}[s]
	if !err {
		return fmt.Errorf("invalid TransmissionStatus %q", s)
	}
	*as = got
	return nil
}
func IsTransmissionStatus(as string) bool {
	_, err := map[string]TransmissionStatus{"FAILED": Failed, "SENT": Sent, "ACKNOWLEDGED": Acknowledged, "TRXESCALATED": Trxescalated}[as]
	if !err {
		return false
	}
	return true
}
