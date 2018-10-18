/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
package scheduler

const (
	/* -------------- Constants for Command -------------------- */
	ID                       string = "id"
	NAME                     string = "name"
	YAML                     string = "yaml"
	COMMAND                  string = "command"
	KEY                      string = "key"
	VALUE                    string = "value"
	TIMELAYOUT                      = "20060102T150405"
	UNLOCKED string = "UNLOCKED"
	ENABLED  string = "ENABLED"

	ContentTypeKey       = "Content-Type"
	ContentTypeJsonValue = "application/json; charset=utf-8"
	ContentLengthKey     = "Content-Length"
)
