#  ----------------------------------------------------------------------------------
#  Copyright 2018 ForgeRock AS.
#  Copyright 2019 Dell Technologies, Inc.
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
#  @author: Alain Pulluelo, ForgeRock (created: June 14, 2018)
#  @version: 1.0.0
#
#  @author: Trevor Conn, Dell Technologies, Inc. (created: July 2, 2019)
#  @version: 1.0.1
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

FROM vault:1.0.2

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2019: Dell Technologies, Inc.'

USER root

# Vault Config File
WORKDIR /vault/config
COPY --from=build-env /edgex-go/cmd/security-secrets-setup/res/local-tls.hcl ./local.hcl

WORKDIR /vault
RUN mkdir res

# Vault PKI/TLS setup/config binary
COPY --from=build-env /edgex-go/cmd/security-secrets-setup/security-secrets-setup .
# Service configuration file
COPY --from=build-env /edgex-go/cmd/security-secrets-setup/res/configuration.toml ./res
# Vault PKI/TLS materials
COPY --from=build-env /edgex-go/cmd/security-secrets-setup/res/pkisetup-vault.json ./res
# Kong PKI/TLS materials
COPY --from=build-env /edgex-go/cmd/security-secrets-setup/res/pkisetup-kong.json ./res

# Create assets folder (needed for unseal key/s, root token and tmp)
# Run CA/Vault and Kong PKI/TLS setups and peform housekeeping tasks
RUN mkdir /vault/config/assets && \
    chown -R vault:vault /vault && \
    chmod 644 /vault/config/local.hcl && \
    chmod 744 security-secrets-setup && \
    /vault/security-secrets-setup --config /vault/res/pkisetup-vault.json && \
    echo "" && \
    /vault/security-secrets-setup --config /vault/res/pkisetup-kong.json && \
    chown -R vault:vault /vault/config/pki

VOLUME /vault/config