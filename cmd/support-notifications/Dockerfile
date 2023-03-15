#  ----------------------------------------------------------------------------------
#  Copyright 2018 Cavium
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

ARG BUILDER_BASE=golang:1.20-alpine3.17
FROM ${BUILDER_BASE} AS builder

ARG ADD_BUILD_TAGS=""

WORKDIR /edgex-go

RUN apk add --update --no-cache make bash git ca-certificates

COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .
RUN make cmd/support-notifications/support-notifications

FROM alpine:3.17

RUN apk add --update --no-cache ca-certificates dumb-init

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2018: Cavium, Copyright (c) 2023: Intel Corporation'

ENV APP_PORT=59860
#expose support notifications port
EXPOSE $APP_PORT

COPY --from=builder /edgex-go/Attribution.txt /
COPY --from=builder /edgex-go/security.txt /
COPY --from=builder /edgex-go/cmd/support-notifications/support-notifications /
COPY --from=builder /edgex-go/cmd/support-notifications/res/configuration.yaml /res/configuration.yaml

ENTRYPOINT ["/support-notifications"]
CMD ["-cp=consul.http://edgex-core-consul:8500", "--registry"]
