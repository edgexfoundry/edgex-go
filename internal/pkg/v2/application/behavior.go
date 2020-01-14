/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

package application

// Behavior defines the common structure used to represent a supported behavior; version specifies the
// version of the behavior (e.g. "2"), kind specifies the type of the behavior (e.g. "ping"), and action specifies a
// transport-agnostic verb (e.g. "read") associated with the behavior.
type Behavior struct {
	Version string
	Kind    string
	Action  string
}

// NewBehavior is a factory function that returns a Behavior struct.
func NewBehavior(version, kind, action string) *Behavior {
	return &Behavior{
		Version: version,
		Kind:    kind,
		Action:  action,
	}
}

// NewBehaviorSlice is a factory function that returns a slice of behavior that contains a single behavior element.
func NewBehaviorSlice(version, kind, action string) []*Behavior {
	return []*Behavior{
		NewBehavior(version, kind, action),
	}
}
