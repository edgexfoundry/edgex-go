/*******************************************************************************
 * Copyright 2021 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package helper

import (
	"fmt"
	"io"
	"text/template"
)

/* Redis ACL configuration
*
* The followings are some contents excerpted from Redis 6.0 online conf documentation
* regarding the ACL rules as seen in this https://redis.io/topics/acl#acl-rules:
*
# Redis ACL users are defined in the following format:
#
#   user <username> ... acl rules ...
#
# For example:
#
#   user worker +@list +@connection ~jobs:* on >ffa9203c493aa99
#
# The special username "default" is used for new connections. If this user
# has the "nopass" rule, then new connections will be immediately authenticated
# as the "default" user without the need of any password provided via the
# AUTH command. Otherwise if the "default" user is not flagged with "nopass"
# the connections will start in not authenticated state, and will require
# AUTH (or the HELLO command AUTH option) in order to be authenticated and
# start to work.
#
# The ACL rules that describe what a user can do are the following:
#
#  on           Enable the user: it is possible to authenticate as this user.
#  +@<category> Allow the execution of all the commands in such category
#               with valid categories are like @admin, @set, @sortedset, ...
#               and so forth, see the full list in the server.c file where
#               the Redis command table is described and defined.
#               The special category @all means all the commands, but currently
#               present in the server, and that will be loaded in the future
#               via modules.
#  allcommands  Alias for +@all. Note that it implies the ability to execute
#               all the future commands loaded via the modules system.
#  allkeys      Alias for ~*
#  ><password>  Add this password to the list of valid password for the user.
#               For example >mypass will add "mypass" to the list.
#               This directive clears the "nopass" flag (see later).
#
# ACL rules can be specified in any order: for instance you can start with
# passwords, then flags, or key patterns. However note that the additive
# and subtractive rules will CHANGE MEANING depending on the ordering.
#
# Basically ACL rules are processed left-to-right.
#
# For more information about ACL configuration please refer to
# the Redis web site at https://redis.io/topics/acl
#
# IMPORTANT NOTE: starting with Redis 6 "requirepass" is just a compatibility
# layer on top of the new ACL system. The option effect will be just setting
# the password for the default user. Clients will still authenticate using
# AUTH <password> as usually, or more explicitly with AUTH default <password>
# if they follow the new protocol: both will work.
#
# requirepass foobared
*
* For EdgeX's use case today, the ACL rules are defined in a template as seen in
* the following constant aclDefaultUserTemplate.
* In particular, it defines the ACL rule for default user with the following accesses:
*   1) allkeys: it allows the user to access all the keys
*   2) +@all: this is an alias for allcommands and + means to allow
*   3) -@dangerous: disallow all the commands that are tagged as dangerous inside the Redis command table
*   4) >{{.RedisPwd}}: add the dynamically injected password for this user
*
*
*/

const (
	// aclDefaultUserTemplate is the ACL rule for "default" user
	aclDefaultUserTemplate = "user default on allkeys +@all -@dangerous >{{.RedisPwd}}"

	// requirePassTemplate is the authenticate password for "default" user
	requirePassTemplate = "requirepass {{.RedisPwd}}"
)

// GenerateConfig writes the redis config based on the pre-defined templates
func GenerateConfig(wr io.Writer, pwd *string) error {
	acl, err := template.New("redis-acl").Parse(aclDefaultUserTemplate + fmt.Sprintln())
	if err != nil {
		return fmt.Errorf("failed to parse ACL template %s: %v", aclDefaultUserTemplate, err)
	}

	// writing the ACL rules:
	if err := acl.Execute(wr, map[string]interface{}{
		"RedisPwd": pwd,
	}); err != nil {
		return fmt.Errorf("failed to execute ACL for config %s: %v", aclDefaultUserTemplate, err)
	}

	// writing the required pwd:
	requirePass, err := template.New("redis-require-pass").Parse(requirePassTemplate + fmt.Sprintln())
	if err != nil {
		return fmt.Errorf("failed to parse requirePass template %s: %v", requirePassTemplate, err)
	}

	if err := requirePass.Execute(wr, map[string]interface{}{
		"RedisPwd": pwd,
	}); err != nil {
		return fmt.Errorf("failed to execute requirePass for config %s: %v", requirePassTemplate, err)
	}

	return nil
}
