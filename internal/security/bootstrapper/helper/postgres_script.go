/*******************************************************************************
 * Copyright (C) 2024-2025 IOTech Ltd
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
	"errors"
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
  SELECT 'CREATE DATABASE edgex_db' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'edgex_db')\gexec
  SELECT 'CREATE GROUP edgex_user' WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'edgex_user') \gexec
  GRANT CONNECT, CREATE ON DATABASE edgex_db TO edgex_user;
  \connect edgex_db;

  DO $$
  BEGIN
    {{range .Services}} 
        CREATE SCHEMA IF NOT EXISTS "{{.Username}}";
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '{{.Username}}') THEN
            CREATE USER "{{.Username}}" with PASSWORD '{{.Password}}';
            GRANT ALL ON SCHEMA "{{.Username}}" TO "{{.Username}}";
            ALTER GROUP edgex_user ADD USER "{{.Username}}";
        END IF;
	{{end}}
  END $$;
EOSQL`
)

// GeneratePostgresScript writes the initialize Postgres db script files based on pre-defined template
// to create the edgex_db and multiple users/password for different EdgeX services
func GeneratePostgresScript(confFile *os.File, credMap []map[string]any) error {
	scriptFile, err := template.New("postgres-script").Parse(scriptTemplate + fmt.Sprintln())
	if err != nil {
		return fmt.Errorf("failed to parse postgres script template: %v", err)
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

// GeneratePasswordFile creates a random password and writes it to the Postgres password file
func GeneratePasswordFile(confFile *os.File, password string) error {
	if password == "" {
		return errors.New("failed to GeneratePasswordFile: password is empty")
	}

	_, err := confFile.WriteString(password)
	return err
}
