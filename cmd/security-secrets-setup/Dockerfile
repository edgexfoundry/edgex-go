#  ----------------------------------------------------------------------------------
#  Copyright 2018 ForgeRock AS.
#  Copyright 2019 Dell Technologies, Inc.
#  Copyright 2019 Intel Corp.
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
#  @author: Alain Pulluelo, ForgeRock (created: Jun 14, 2018)
#
#  @author: Trevor Conn, Dell Technologies, Inc. (created: July 2, 2019)
#
#  @author: Jim Wang, Intel Corp. (modified: August 8, 2019)
#
#  SPDX-License-Identifier: Apache-2.0'
#  ----------------------------------------------------------------------------------

# Docker image for building EdgeX Foundry security-secrets-setup
FROM golang:1.13-alpine AS build-env

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

FROM alpine:latest

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2019: Dell Technologies, Inc.  Copyright (c) 2019 Intel Corporation'

USER root

# install necessary tools
RUN apk update && apk add ca-certificates dumb-init \
    && rm -rf /var/cache/apk/* 

ENV BASE_DIR=/edgex/security-secrets-setup

WORKDIR $BASE_DIR
RUN mkdir res && \
    mkdir -p /vault/init && \
    mkdir /vault/staging

# Vault PKI/TLS setup/config binary
COPY --from=build-env /edgex-go/cmd/security-secrets-setup/security-secrets-setup .
# Service configuration file
COPY --from=build-env /edgex-go/cmd/security-secrets-setup/res/configuration.toml ./res
# Vault PKI/TLS materials
COPY --from=build-env /edgex-go/cmd/security-secrets-setup/res/pkisetup-vault.json ./res
# Kong PKI/TLS materials
COPY --from=build-env /edgex-go/cmd/security-secrets-setup/res/pkisetup-kong.json ./res

# Vault's startup script
COPY --from=build-env /edgex-go/cmd/security-secrets-setup/start_vault.sh /vault/staging
RUN chmod +x /vault/staging/start_vault.sh

# setup the entry point script
# it uses `dumb-init` as the top-level process to reap any
# zombie processes created by sub-processes
COPY --from=build-env /edgex-go/cmd/security-secrets-setup/entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh \
    && ln -s /usr/local/bin/entrypoint.sh /

RUN chmod 755 security-secrets-setup

ENTRYPOINT ["entrypoint.sh"]

CMD [ "generate" ]
