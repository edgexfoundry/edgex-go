//
// Copyright (c) 2020 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

@Library("edgex-global-pipelines@e3a708771a572f3309a02b78ef07992b923c5409") _

edgeXBuildGoParallel(
    project: 'edgex-go',
    dockerFileGlobPath: 'cmd/**/Dockerfile',
    testScript: 'make test',
    buildScript: 'make build',
    publishSwaggerDocs: false,
    swaggerApiFolders: ['openapi/v1', 'openapi/v2'],
    buildSnap: true
)
