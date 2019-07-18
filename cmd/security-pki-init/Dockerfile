#  ----------------------------------------------------------------------------------
#  Copyright (c) 2019 Intel Corporation
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#
#  SPDX-License-Identifier: Apache-2.0'
#  ----------------------------------------------------------------------------------
# Docker image for building EdgeX Foundry Config Seed
FROM golang:1.12-alpine AS build-env

# environment variables
ENV GO111MODULE=on

# set the working directory
WORKDIR /edgex-go

RUN apk update && apk add make git

# copy go source files
COPY go.mod .

RUN go mod download

COPY . .

RUN make cmd/security-secrets-setup/security-secrets-setup

RUN make cmd/security-pki-init/security-pki-init

# the pki-init base image starts here
FROM alpine:latest

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2019 Intel Corporation'

ENV PKI_INIT_DIR=/edgex/security-pki-init

ENV PKI_CACHE=/etc/edgex/pki

RUN set -eux; \
    apk add --no-cache ca-certificates dumb-init

RUN mkdir -p $PKI_INIT_DIR

WORKDIR $PKI_INIT_DIR

# setup / install the executable
COPY --from=build-env /edgex-go/cmd/security-pki-init/security-pki-init .
COPY --from=build-env /edgex-go/cmd/security-secrets-setup/security-secrets-setup .

# setup configuration json file for vault
COPY --from=build-env /edgex-go/cmd/security-secrets-setup/res/pkisetup-vault.json .

# resource files
COPY --from=build-env /edgex-go/cmd/security-pki-init/res/ ./res/

# setup the entry point script
# it uses `dumb-init` as the top-level process to reap any
# zombie processes created by sub-processes
COPY --from=build-env /edgex-go/cmd/security-pki-init/docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
RUN ln -s usr/local/bin/docker-entrypoint.sh /  # backwards compatability

ENTRYPOINT ["docker-entrypoint.sh"]

CMD ["sh", "-c", "${PKI_INIT_DIR}/security-pki-init -generate"]
