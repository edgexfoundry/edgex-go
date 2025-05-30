#  ----------------------------------------------------------------------------------
#  Copyright 2023 Intel Corporation
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

# Build utility container
ARG BUILDER_BASE=golang:1.23-alpine3.20
FROM ${BUILDER_BASE} AS builder

FROM ghcr.io/spiffe/spire-server:1.11.1 as spire_server
FROM ghcr.io/spiffe/spire-agent:1.11.1 as spire_agent

# Deployment image
FROM alpine:3.20

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2022 Intel Corporation'

RUN apk update && apk --no-cache --update add dumb-init openssl gcompat
# Ensure using latest versions of all installed packages to avoid any recent CVEs
RUN apk --no-cache upgrade

COPY --from=spire_agent /opt/spire/bin/spire-agent /usr/local/bin
COPY --from=spire_server /opt/spire/bin/spire-server /usr/local/bin

COPY Attribution.txt /
COPY security.txt /

COPY cmd/security-spire-agent/docker-entrypoint.sh /usr/local/bin/
COPY cmd/security-spire-agent/agent.conf /usr/local/etc/spire/agent.conf.tpl
COPY cmd/security-spire-agent/openssl.conf /usr/local/etc/

ENTRYPOINT [ "/usr/bin/dumb-init" ]
CMD [ "--verbose", "docker-entrypoint.sh" ]
