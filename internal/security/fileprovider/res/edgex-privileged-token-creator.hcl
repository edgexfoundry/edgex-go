//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//

//
// This file is taken from
// https://raw.githubusercontent.com/edgexfoundry/edgex-docs/master/security/token-file-provider.1.rst
//
// This is a reference copy of the policy that security-file-token-provider requires
// in order to run and create policies for other services.
//

path "auth/token/create" {
  capabilities = ["create", "update", "sudo"]
}

path "auth/token/create-orphan" {
  capabilities = ["create", "update", "sudo"]
}

path "auth/token/create/*" {
  capabilities = ["create", "update", "sudo"]
}

path "sys/policies/acl/edgex-service-*"
{
  capabilities = ["create", "read", "update", "delete" ]
}

path "sys/policies/acl"
{
  capabilities = ["list"]
}