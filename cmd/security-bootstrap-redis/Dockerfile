#########
# Copyright 2020 Redis Labs
#
# Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
# in compliance with the License. You may obtain a copy of the License at
# 
# http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software distributed under the License
# is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
# or implied. See the License for the specific language governing permissions and limitations under
# the License.
# 
# @author: Andre Srinivasan
#########

ARG BUILDER_BASE=golang:1.15-alpine
FROM ${BUILDER_BASE} AS builder

WORKDIR /edgex-go

# The main mirrors are giving us timeout issues on builds periodically.
# So we can try these.

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

RUN apk update && apk add make git

COPY go.mod .

RUN go mod download

COPY . .

RUN make cmd/security-bootstrap-redis/security-bootstrap-redis

FROM alpine:3.10

RUN apk update && apk add ca-certificates dumb-init curl \
    && rm -rf /var/cache/apk/* 

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2020 Redis Labs'

WORKDIR /
COPY --from=builder /edgex-go/cmd/security-bootstrap-redis/Attribution.txt /
COPY --from=builder /edgex-go/cmd/security-bootstrap-redis/security-bootstrap-redis /
COPY --from=builder /edgex-go/cmd/security-bootstrap-redis/entrypoint.sh /
COPY --from=builder /edgex-go/cmd/security-bootstrap-redis/res/configuration.toml /res/configuration.toml
RUN chmod +x entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
