#  ----------------------------------------------------------------------------------
#  Copyright (C) 2024-2025 IOTech Ltd
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

ARG BUILDER_BASE=golang:1.25-alpine3.22
FROM ${BUILDER_BASE} AS builder

ARG ADD_BUILD_TAGS=""

WORKDIR /edgex-go

RUN apk add --update --no-cache make bash git ca-certificates

COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .
RUN make cmd/support-scheduler/support-scheduler

FROM alpine:3.23.2

RUN apk add --update --no-cache ca-certificates dumb-init tzdata
# Ensure using latest versions of all installed packages to avoid any recent CVEs
RUN apk --no-cache upgrade

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2024: IOTech'

ENV APP_PORT=59863
#expose support scheduler port
EXPOSE $APP_PORT

COPY --from=builder /edgex-go/Attribution.txt /
COPY --from=builder /edgex-go/security.txt /
COPY --from=builder /edgex-go/cmd/support-scheduler/support-scheduler /
COPY --from=builder /edgex-go/cmd/support-scheduler/res /res

ENTRYPOINT ["/support-scheduler"]
CMD ["-cp=keeper.http://edgex-core-keeper:59890", "--registry"]
