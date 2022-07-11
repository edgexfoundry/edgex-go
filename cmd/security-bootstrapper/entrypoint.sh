#!/usr/bin/dumb-init /bin/sh
#
# The entry point script uses dumb-init as the top-level process to reap any
# zombie processes
#
#  ----------------------------------------------------------------------------------
#  Copyright (c) 2022 Intel Corporation
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

set -e

# Passing the arguments to the executable as $@ contains only the CMD arguments without the executable name
# treat anything not /bin/sh as to run this security-bootstrapper executable with the arguments
# this is useful for debugging the container like running with `docker run -it --rm security-bootstrapper /bin/sh`
if [ ! "$1" = '/bin/sh' ]; then
    set -- security-bootstrapper "$@"
fi

DEFAULT_EDGEX_USER_ID=2002
EDGEX_USER_ID=${EDGEX_USER:-$DEFAULT_EDGEX_USER_ID}

# assumming the target directory ${SECURITY_INIT_DIR} has been created by the framework
cp -rpd ${SECURITY_INIT_STAGING}/* ${SECURITY_INIT_DIR}/

# During the bootstrapping, environment variables come for compose file environment files,
# which then injecting into all other related containers on other services' entrypoint scripts
# if the executable is not 'security-bootstrapper'; then we consider it not running the bootstrapping process
# for the user may just want to debug into the container shell itself
if [ "$1" = 'security-bootstrapper' ]; then
  # run the executable as ${EDGEX_USER}
  echo "$(date) Executing ./$@"
  exec su-exec ${EDGEX_USER_ID} "./$@"

else
  # for debug purposes like docker run -it --rm security-bootstrapper:0.0.0-dev /bin/sh
  echo "current directory:" "$PWD"
  exec su-exec ${EDGEX_USER_ID} "$@"
fi