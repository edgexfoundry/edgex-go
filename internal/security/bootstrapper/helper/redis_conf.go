/*******************************************************************************
 * Copyright 2023 Intel Corporation
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
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
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
#
#
# Using an external ACL file
#
# Instead of configuring users here in this file, it is possible to use
# a stand-alone file just listing users. The two methods cannot be mixed:
# if you configure users here and at the same time you activate the external
# ACL file, the server will refuse to start.
#
# The format of the external ACL user file is exactly the same as the
# format that is used inside redis.conf to describe users.
#
# aclfile /etc/redis/users.acl
#
*
* For EdgeX's use case today, the ACL rules are defined in a template as seen in
* the following constant aclDefaultUserTemplate.
* In particular, it defines the ACL rule for default user with the following accesses:
*   1) allkeys: it allows the user to access all the keys
*   2) +@all: this is an alias for allcommands and + means to allow
*   3) -@dangerous: disallow all the commands that are tagged as dangerous inside the Redis command table
*   4) #{{.Sha256RedisPwd}}: add the dynamically injected SHA-256 hashed password for this user
*
* We do not use "requirepass" directive any more since hashed password already specified in the ACL rule
* for that user.
*
*/

const (
	// the default user name from redis built-in
	redisDefaultUser = "default"

	// aclFileConfigTemplate is the external acl file for redis config
	aclFileConfigTemplate = `aclfile {{.ACLFilePath}}
maxclients {{.MaxClients}}
`

	// aclDefaultUserTemplate is the ACL rule for "default" user
	aclDefaultUserTemplate = "user {{.RedisUser}} on allkeys allchannels +@all -@dangerous #{{.Sha256RedisPwd}}"
)

// GenerateRedisConfig writes the startup configuration of Redis server based on pre-defined template
func GenerateRedisConfig(confFile *os.File, aclfilePath string, maxClients int) error {
	if maxClients <= 0 {
		return fmt.Errorf("number of maxClient should be greater than 0 but found %d", maxClients)
	}

	aclfile, err := template.New("redis-conf").Parse(aclFileConfigTemplate + fmt.Sprintln())
	if err != nil {
		return fmt.Errorf("failed to parse Redis conf template %s: %v", aclFileConfigTemplate, err)
	}

	// writing the config file
	fwriter := bufio.NewWriter(confFile)
	if err := aclfile.Execute(fwriter, map[string]interface{}{
		"ACLFilePath": aclfilePath,
		"MaxClients":  maxClients,
	}); err != nil {
		return fmt.Errorf("failed to execute external ACL file for config %s: %v", aclFileConfigTemplate, err)
	}

	if err := fwriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush the config file writer buffer %v", err)
	}

	return nil
}

// GenerateACLConfig writes the redis ACL file based on the pre-defined templates
func GenerateACLConfig(aclFile *os.File, pwd *string) error {
	// the metadata for Redis ACL file
	type redisACL struct {
		RedisUser      string
		Sha256RedisPwd string
	}

	acl, err := template.New("redis-acl").Parse(aclDefaultUserTemplate + fmt.Sprintln())
	if err != nil {
		return fmt.Errorf("failed to parse ACL template %s: %v", aclDefaultUserTemplate, err)
	}

	hashed256 := sha256.Sum256([]byte(*pwd))

	// writing the ACL rules:
	fwriter := bufio.NewWriter(aclFile)
	if err := acl.Execute(fwriter, redisACL{
		RedisUser:      redisDefaultUser,
		Sha256RedisPwd: fmt.Sprintf("%x", hashed256),
	}); err != nil {
		return fmt.Errorf("failed to execute ACL for config %s: %v", aclDefaultUserTemplate, err)
	}

	if err := fwriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush the ACL file writer buffer %v", err)
	}

	return nil
}
