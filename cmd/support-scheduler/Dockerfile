#  ----------------------------------------------------------------------------------
#  Copyright 2018 Dell Technologies, Inc.
#  Copyright 2018 Cavium
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

ARG BUILDER_BASE=golang:1.18-alpine3.16
FROM ${BUILDER_BASE} AS builder

ARG ADD_BUILD_TAGS=""

WORKDIR /edgex-go

RUN apk add --update --no-cache make git

COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .
RUN make cmd/support-scheduler/support-scheduler

FROM alpine:3.16

RUN apk add --update --no-cache dumb-init

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2018: Dell, Cavium, Copyright (c) 2021: Intel Corporation'

ENV APP_PORT=59861
#expose support scheduler port
EXPOSE $APP_PORT

COPY --from=builder /edgex-go/Attribution.txt /
COPY --from=builder /edgex-go/cmd/support-scheduler/support-scheduler /
COPY --from=builder /edgex-go/cmd/support-scheduler/res/configuration.toml /res/configuration.toml

ENTRYPOINT ["/support-scheduler"]
CMD ["-cp=consul.http://edgex-core-consul:8500", "--registry", "--confdir=/res"]
