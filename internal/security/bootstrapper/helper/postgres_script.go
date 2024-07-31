/*******************************************************************************
 * Copyright (C) 2024 IOTech Ltd
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
	"fmt"
	"os"
	"text/template"
)

const (
	ServicesTempVarName = "Services"
	UsernameTempVarName = "Username"
	PasswordTempVarName = "Password"

	// file template for init-db script file
	scriptTemplate = `#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" -q <<-'EOSQL'
  CREATE DATABASE edgex_db;

  DO $$
  BEGIN
    {{range .%s}}
		CREATE USER "{{.%s}}" with PASSWORD '{{.%s}}';
	{{end}}
  END $$;
EOSQL`
)

// GeneratePostgresScript writes the initialize Postgres db script files based on pre-defined template
// to create the edgex_db and multiple users/password for different EdgeX services
func GeneratePostgresScript(confFile *os.File, credMap []map[string]any) error {
	finalScriptTemplate := fmt.Sprintf(scriptTemplate, ServicesTempVarName, UsernameTempVarName, PasswordTempVarName)
	scriptFile, err := template.New("postgres-script").Parse(finalScriptTemplate + fmt.Sprintln())
	if err != nil {
		return fmt.Errorf("failed to parse Redis conf template %s: %v", aclFileConfigTemplate, err)
	}

	// writing the config file
	fWriter := bufio.NewWriter(confFile)
	if err := scriptFile.Execute(fWriter, map[string]any{
		ServicesTempVarName: credMap,
	}); err != nil {
		return fmt.Errorf("failed to execute Postgres init-db templat: %v", err)
	}

	if err := fWriter.Flush(); err != nil {
		return fmt.Errorf("failed to flush the config file writer buffer %v", err)
	}

	return nil
}
