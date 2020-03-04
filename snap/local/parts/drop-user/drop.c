/* -*- Mode: C; indent-tabs-mode: nil; tab-width: 4 -*-
 *
 * Copyright (C) 2020 Canonical Ltd
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
 */

#define _GNU_SOURCE
#include <dlfcn.h>
#include <err.h>
#include <errno.h>
#include <pwd.h>
#include <grp.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>

/* This code is based on example code published on launchpad:
 *
 * - https://git.launchpad.net/~jdstrand/+git/test-setgroups
 *
 * The Fuji snap originally used gosu command to run postgres commands as the
 * 'snap_daemon' user, but as gosu doesn't support the extrausers passwd db
 * extension used on Ubuntu Core, the snap couldn't be installed on a Core system.
*/

static int (*original_setgroups) (size_t, const gid_t[]);

int main(int argc, char *argv[])
{
	char **cmdargv;
	char *user = "snap_daemon";

	if (argc < 2) {
	  printf("Usage: %s command [args]\n", argv[0]);
	  exit(0);
	}

	cmdargv = &argv[1];

	original_setgroups = dlsym(RTLD_NEXT, "setgroups");
	if (!original_setgroups) {
		fprintf(stderr, "could not find setgroups in libc; %s", dlerror());
		return -1;
	}

	/* Convert our username to a passwd entry */
	struct passwd *pwd = getpwnam(user);
	if (pwd == NULL) {
		printf("'%s' not found\n", user);
		exit(EXIT_FAILURE);
	}

	/* Drop supplementary groups first if can, using portable method
	 * (should fail without LD_PRELOAD)
	 */
	if (geteuid() == 0 && original_setgroups(0, NULL) < 0) {
		perror("setgroups");
		goto fail;
	}

	/* Drop gid after supplementary groups */
	if (setgid(pwd->pw_gid) < 0) {
		perror("setgid");
		goto fail;
	}

	/* Drop uid after gid */
	if (setuid(pwd->pw_uid) < 0) {
		perror("setuid");
		goto fail;
	}

	execvp(cmdargv[0], cmdargv);
	err(1, "%s", cmdargv[0]);
	exit (1);

 fail:
	exit(EXIT_FAILURE);
}
