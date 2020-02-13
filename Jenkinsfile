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

edgeXGeneric([
    project: 'edgex-go',
    mavenSettings: ['edgex-go-settings:SETTINGS_FILE', 'edgex-go-codecov-token:CODECOV_TOKEN'],
    credentials: [ string(credentialsId: 'swaggerhub-api-key', variable: 'APIKEY') ],
    env: [
        GOPATH: '/opt/go-custom/go',
        GO_VERSION: '1.13',
        REPO_ROOT: '$HOME/$BUILD_ID/gopath/src/github.com/edgexfoundry/edgex-go/',
        DEPLOY_TYPE: 'staging'
    ],
    path: [
        '/opt/go-custom/go/bin'
    ],
    branches: [
        '*': [
            pre_build: ['shell/install_custom_golang.sh'],
            build: [
                'make test raml_verify && make build docker',
                'shell/codecov-uploader.sh'
            ]
        ],
        'master': [
            post_build: [ 'shell/edgexfoundry-go-docker-push.sh', 'shell/edgex-publish-swagger.sh' ]
        ]
    ]
])
