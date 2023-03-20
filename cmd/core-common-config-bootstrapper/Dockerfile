#  ----------------------------------------------------------------------------------
#  Copyright (c) 2023 Intel Corporation
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
#  SPDX-License-Identifier: Apache-2.0
#  ----------------------------------------------------------------------------------

# Docker image for Golang Core common config bootstrapper service
ARG BUILDER_BASE=golang:1.20-alpine3.17
FROM ${BUILDER_BASE} AS builder

ARG ADD_BUILD_TAGS=""

WORKDIR /edgex-go

RUN apk add --update --no-cache make git

COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .
RUN make -e ADD_BUILD_TAGS=$ADD_BUILD_TAGS cmd/core-common-config-bootstrapper/core-common-config-bootstrapper

#Next image - Copy built Go binary into new workspace
FROM alpine:3.17

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2023: Intel Corporation'


RUN apk add --update --no-cache dumb-init
COPY --from=builder /edgex-go/Attribution.txt /
COPY --from=builder /edgex-go/security.txt /
COPY --from=builder /edgex-go/cmd/core-common-config-bootstrapper/core-common-config-bootstrapper /
COPY --from=builder /edgex-go/cmd/core-common-config-bootstrapper/res/configuration.yaml /res/configuration.yaml

COPY --from=builder /edgex-go/cmd/core-common-config-bootstrapper/entrypoint.sh /usr/local/bin/
RUN chmod 755 /usr/local/bin/entrypoint.sh \
    && ln -s /usr/local/bin/entrypoint.sh /

ENTRYPOINT ["entrypoint.sh"]
CMD ["/core-common-config-bootstrapper", "-cp=consul.http://edgex-core-consul:8500"]
