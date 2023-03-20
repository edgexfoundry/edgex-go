#  ----------------------------------------------------------------------------------
#  Copyright 2019 Dell Technologies, Inc.
#  Copyright 2020-2023 Intel Corporation
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
#  ----------------------------------------------------------------------------------

ARG BUILDER_BASE=golang:1.20-alpine3.17
FROM ${BUILDER_BASE} AS builder

WORKDIR /edgex-go

RUN apk add --update --no-cache make git

COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .
RUN make cmd/secrets-config/secrets-config

FROM alpine:3.17

RUN apk add --update --no-cache dumb-init openssl

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2019: Dell Technologies, Inc.; Copyright (C) 2023 Intel Corporation'

WORKDIR /edgex

COPY --from=builder /edgex-go/Attribution.txt /
COPY --from=builder /edgex-go/security.txt /

# Note that secrets-config shares the same configuration file as security-proxy-setup
# as we are splitting security-proxy-setup into two different utilities for ease-of-use.
COPY --from=builder /edgex-go/cmd/secrets-config/res/configuration.yaml res/configuration.yaml
COPY --from=builder /edgex-go/cmd/secrets-config/secrets-config .

# Setup the entry point script and assign perms
COPY --from=builder /edgex-go/cmd/security-proxy-setup/entrypoint.sh /usr/local/bin/
RUN chmod 755 /usr/local/bin/entrypoint.sh \
    && ln -s /usr/local/bin/entrypoint.sh /

ENTRYPOINT ["entrypoint.sh"]
