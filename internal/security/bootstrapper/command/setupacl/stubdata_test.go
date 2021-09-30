/*******************************************************************************
 * Copyright 2021 Intel Corporation
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

package setupacl

// this is just the stub for test data related
const (
	// nolint:gosec
	secretstoreTokenJsonStub = `
	{
		"auth": {
		  "accessor": "xxxxxxxxxxxxxxxxxxxxxxx",
		  "client_token": "yyyyyyyyyyyyyyyyyyyyyyyyyy",
		  "entity_id": "",
		  "lease_duration": 3600,
		  "metadata": {
			"description": "Consul secrets engine management token"
		  },
		  "orphan": true,
		  "policies": [
			"consul_secrets_engine_management_policy",
			"default"
		  ],
		  "renewable": true,
		  "token_policies": [
			"consul_secrets_engine_management_policy",
			"default"
		  ],
		  "token_type": "service"
		},
		"data": null,
		"lease_duration": 0,
		"lease_id": "",
		"renewable": false,
		"request_id": "aaaaaaaa-1111-2222-bbbb-cccccccccccc",
		"warnings": null,
		"wrap_info": null
	}
	`
)
