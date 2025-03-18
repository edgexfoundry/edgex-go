//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package embed

import "embed"

// SQLFiles contains the SQL files as embedded resources.
// Following code use go embed directive to embed the SQL files into the binary.

//go:embed sql
var SQLFiles embed.FS

// The SQL files are stored in the sql directory with two subdirectories: idempotent and versions.
// 1. idempotent: directory contains the SQL files that can be initialized the db schema.
//    The SQL files in this directory are designed to be idempotent and can be executed multiple times without changing
//    the result.
// 2. versions: directory contains various version subdirectories with the SQL files that are used to update table
//    schema per versions.
//
// When any future requirements need to alter the table schema, the practice is to AVOID directly update SQL files in
// idempotent directory. Instead, create a new subdirectory with the new semantic version number. Add new SQL files to
// update the schema into the new version subdirectory. The SQL files in the new version subdirectory should be named
// with the format of <execution_order>-<description>.sql. Moreover, when naming the new version subdirectory, follow
// the semantic versioning rules as defined in https://semver.org/#backusnaur-form-grammar-for-valid-semver-versions.
// The valid semver format is <valid semver> ::= <version core> "-" <pre-release>, so use -dev rather than .dev as
// pre-release suffix for semver to parse correctly.
