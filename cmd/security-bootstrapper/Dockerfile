#  ----------------------------------------------------------------------------------
#  Copyright 2023 Intel Corp.
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
#
#  SPDX-License-Identifier: Apache-2.0
#  ----------------------------------------------------------------------------------

ARG BUILDER_BASE=golang:1.20-alpine3.17
FROM ${BUILDER_BASE} AS builder

WORKDIR /edgex-go

RUN apk add --update --no-cache make git

COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .
RUN make cmd/security-bootstrapper/security-bootstrapper

FROM alpine:3.17

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2023 Intel Corporation'

RUN apk add --update --no-cache dumb-init su-exec

ENV SECURITY_INIT_DIR /edgex-init
ENV SECURITY_INIT_STAGING /edgex-init-staging
ARG BOOTSTRAP_REDIS_DIR=${SECURITY_INIT_STAGING}/bootstrap-redis
ARG BOOTSTRAP_MOSQUITTO_DIR=${SECURITY_INIT_STAGING}/bootstrap-mosquitto

RUN mkdir -p ${BOOTSTRAP_REDIS_DIR} ${BOOTSTRAP_MOSQUITTO_DIR}

WORKDIR ${SECURITY_INIT_STAGING}

# copy all entrypoint scripts into shared folder
COPY --from=builder /edgex-go/cmd/security-bootstrapper/entrypoint-scripts/ ${SECURITY_INIT_STAGING}/
RUN chmod +x ${SECURITY_INIT_STAGING}/*.sh

COPY --from=builder /edgex-go/Attribution.txt /
COPY --from=builder /edgex-go/security.txt /
COPY --from=builder /edgex-go/cmd/security-bootstrapper/security-bootstrapper .
COPY --from=builder /edgex-go/cmd/security-bootstrapper/res/configuration.yaml ./res/

# needed for bootstrapping Redis db
COPY --from=builder /edgex-go/cmd/security-bootstrapper/res-bootstrap-redis/configuration.yaml ${BOOTSTRAP_REDIS_DIR}/res/

# needed for bootstrapping mosquitto
COPY --from=builder /edgex-go/cmd/security-bootstrapper/res-bootstrap-mosquitto/configuration.yaml ${BOOTSTRAP_MOSQUITTO_DIR}/res/

# copy Consul ACL related configs
COPY --from=builder /edgex-go/cmd/security-bootstrapper/consul-acl/ ${SECURITY_INIT_STAGING}/consul-bootstrapper/

# setup entry point script
COPY --from=builder /edgex-go/cmd/security-bootstrapper/entrypoint.sh /
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

# gate is one subcommand for security-bootstrapper to do security bootstrapping
CMD ["gate"]
