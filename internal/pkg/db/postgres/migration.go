//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"embed"
	goErrors "errors"
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

const (
	migrationTableName = "schema_migrations"

	idempotentDBScriptsDir = "sql/idempotent"
	versionMigrationDir    = "sql/versions"

	statusSuccess = "SUCCESS"
	statusFailure = "FAILURE"

	// initialVersion specifies the version when the service introduced schema_migrations table to handle migration
	// as we introduced the schema_migration in the release of 4.0.0-dev, the initial version is set to 4.0.0-dev.
	// NOTE: for semver.* 4.0.0-dev is less than 4.0.0
	initialVersion = "4.0.0-dev"
)

// TableManager defines the interface for the table manager
type TableManager interface {
	RunScripts(ctx context.Context) error
}

// pgTableManager is a struct that implements the TableManager interface for PostgreSQL DB
type pgTableManager struct {
	connPool       *pgxpool.Pool
	logger         logger.LoggingClient
	schemaName     string
	serviceKey     string
	serviceVersion semver.Version
	sqlFiles       embed.FS
}

func NewTableManager(connPool *pgxpool.Pool, lc logger.LoggingClient, schemaName, serviceKey, version string, sqlFiles embed.FS) (TableManager, error) {
	serviceSemver, err := semver.Make(version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service version %s: %w", version, err)
	}
	return &pgTableManager{
		connPool:       connPool,
		logger:         lc,
		schemaName:     schemaName,
		serviceKey:     serviceKey,
		serviceVersion: serviceSemver,
		sqlFiles:       sqlFiles,
	}, nil
}

// RunScripts have the table manager to apply idempotent scripts and migration scripts
func (tm *pgTableManager) RunScripts(ctx context.Context) error {
	// prepare a postgres advisory lock to ensure only one connector can manipulate DB schema at a time
	adLock, lockCreationErr := NewAdvisoryLock(tm.connPool, tm.logger, tm.serviceKey)
	if lockCreationErr != nil {
		return fmt.Errorf("failed to create advisory lock for %s: %w", tm.serviceKey, lockCreationErr)
	}
	locked, err := adLock.Lock()
	if err != nil {
		return fmt.Errorf("error while acquiring advisory lock for %s: %w", tm.serviceKey, err)
	}
	if !locked {
		return fmt.Errorf("could not acquire advisory lock for %s. ensure there are no other connectors "+
			"running and try again", tm.serviceKey)
	}
	defer func() {
		_, unlockErr := adLock.Unlock()
		if unlockErr != nil {
			tm.logger.Errorf("failed to release advisory lock for %s: %s", tm.serviceKey, unlockErr.Error())
		}
	}()

	// run idempotent sql scripts, the service assumes that the idempotent scripts can be executed multiple times
	// without any side effects
	if err = executeSqlFilesDir(ctx, tm.connPool, tm.sqlFiles, idempotentDBScriptsDir); err != nil {
		return fmt.Errorf("%s failed to execute idempotent sql scripts: %w", tm.serviceKey, err)
	}

	// check if schema_migrations table exists or not
	migrationTableExists, err := tm.isMigrationTableAvailable(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to determine if the %s table exists or not: %w", tm.serviceKey,
			migrationTableName, err)
	}

	// if schema_migrations table doesn't exist, create the table
	if !migrationTableExists {
		if err = tm.createMigrationTable(ctx); err != nil {
			return fmt.Errorf("%s failed to create %s table: %w", tm.serviceKey, migrationTableName, err)
		}
	}

	// query the latest record with success status as dbVersion from schema_migrations table
	dbVersion, err := tm.getMostRecentSuccessVersion(ctx)
	if err != nil {
		return fmt.Errorf("%s failed to query the most recent successful version from %s table: %w",
			tm.serviceKey, migrationTableName, err)
	}

	// service version does not include -dev and when using the default value, 0.0.0, it requires some care
	defaultDevVersion, err := semver.Parse("0.0.0")
	if err != nil {
		return fmt.Errorf("%s failed to parse default development version: %w", tm.serviceKey, err)
	}
	// if the table is empty, the dbVersion will be a dummy semver.Version{}, and insert a record for the initial
	// version 4.0.0-dev or the development version 0.0.0-dev, also assign the result to dbVersion.
	if dbVersion.Equals(semver.Version{}) {
		if tm.serviceVersion.Equals(defaultDevVersion) {
			// add -dev so the comparison with 0.0.0 behaves as expected
			dbVersion, err = semver.Parse("0.0.0-dev")
			if err != nil {
				return fmt.Errorf("%s failed to parse development version: %w", tm.serviceKey, err)
			}
		} else {
			dbVersion, err = semver.Parse(initialVersion)
			if err != nil {
				return fmt.Errorf("%s failed to parse initial version: %w", tm.serviceKey, err)
			}
		}

		err = tm.insertVersionRecord(ctx, dbVersion, statusSuccess)
		if err != nil {
			return fmt.Errorf("%s failed to insert initial version: %w", tm.serviceKey, err)
		}
	}

	// check if the dbVersion is above service version
	if dbVersion.Compare(tm.serviceVersion) > 0 {
		return fmt.Errorf("%s detects that the most recent successful version record (%s) in %s table is "+
			"above service version %s", tm.serviceKey, dbVersion.String(), migrationTableName, tm.serviceVersion.String())
	}

	// apply the migration scripts when dbVersion is below service version or when running dev version
	if dbVersion.Compare(tm.serviceVersion) < 0 {
		tm.logger.Infof("db schema version: %s, service version: %s, %s preparing to apply schema migration "+
			"scripts if any", dbVersion.String(), tm.serviceVersion.String(), tm.serviceKey)
		if err = tm.applyMigration(ctx, dbVersion); err != nil {
			return fmt.Errorf("failed to apply migration for %s: %w", tm.serviceKey, err)
		}
	}
	tm.logger.Infof("%s successfully applied SQL scripts", tm.serviceKey)
	return nil
}

func (tm *pgTableManager) applyMigration(ctx context.Context, dbVersion semver.Version) error {
	// get the sorted version folder names
	sortedVersionDirs, err := getSortedVersionDirNames(tm.sqlFiles, versionMigrationDir)
	if err != nil {
		return fmt.Errorf("failed to get sorted version directories: %w", err)
	}

	for _, version := range sortedVersionDirs {
		// skip the version that is less than or equal to the db version
		if version.Compare(dbVersion) <= 0 {
			tm.logger.Debugf("%s skips applying migration scripts for version %s as it is less than or equal to "+
				"the db version %s", tm.serviceKey, version.String(), dbVersion.String())
			continue
		}

		// execute sql files in the version folder as a transaction
		tm.logger.Infof("applying migration scripts for version %s", version.String())
		err = executeSqlFilesDir(ctx, tm.connPool, tm.sqlFiles, fmt.Sprintf("%s/%s", versionMigrationDir, version.String()))
		if err != nil {
			tm.logger.Errorf("%s failed to apply migration scripts for version %s: %s", tm.serviceKey, version.String(), err.Error())
			// if unsuccessful, insert a new record into the schema_migrations table with failure status and return error
			insertErr := tm.insertVersionRecord(ctx, version, statusFailure)
			if insertErr != nil {
				return fmt.Errorf("%s failed to apply migration scripts for version %s with error (%w) and "+
					"also failed to insert failure record into %s table with error (%w)",
					tm.serviceKey, version.String(), err, migrationTableName, insertErr)
			}
			return fmt.Errorf("%s failed to apply migration scripts for version %s: %w", tm.serviceKey, version.String(), err)
		}

		// if successful, insert a new record into the schema_migrations table with success status
		err = tm.insertVersionRecord(ctx, version, statusSuccess)
		if err != nil {
			// the migration scripts have been successfully applied, but failed to insert a record into
			// schema_migrations table, which may result in reapplying the same migration scripts in the next launch of
			// the service
			return fmt.Errorf("%s failed to insert success record for version %s into %s table: %w",
				tm.serviceKey, version.String(), migrationTableName, err)
		}
	}
	return nil
}

func (tm *pgTableManager) insertVersionRecord(ctx context.Context, version semver.Version, status string) error {
	stmt := fmt.Sprintf(`INSERT INTO "%s".%s(version, status) VALUES($1, $2)`, tm.schemaName, migrationTableName)
	_, err := tm.connPool.Exec(ctx, stmt, version, status)
	if err != nil {
		return WrapDBError(fmt.Sprintf("%s failed to insert version %s record into %s table", tm.serviceKey, version.String(), migrationTableName), err)
	}
	return nil
}

func (tm *pgTableManager) isMigrationTableAvailable(ctx context.Context) (exists bool, err error) {
	// check information_schema.tables to see if the schema_migrations table exists
	stmt := fmt.Sprintf("SELECT EXISTS ( SELECT 1 FROM information_schema.tables WHERE table_schema = '%s' AND table_name = '%s' )", tm.schemaName, migrationTableName)
	row := tm.connPool.QueryRow(ctx, stmt)
	err = row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if %s table exists: %w", migrationTableName, err)
	}
	return exists, nil
}

func (tm *pgTableManager) createMigrationTable(ctx context.Context) error {
	// we intentionally create migration table here rather than in every services' idempotent scripts to ensure that
	// schema_migration definition is consistent across all services and only maintained in one place
	createMigrationTableStmt := fmt.Sprintf(`CREATE TABLE "%s".%s (`+
		"id SERIAL PRIMARY KEY,"+
		"version TEXT NOT NULL,"+
		"status TEXT NOT NULL,"+
		"created TIMESTAMPTZ NOT NULL DEFAULT NOW())", tm.schemaName, migrationTableName)
	_, err := tm.connPool.Exec(ctx, createMigrationTableStmt)
	if err != nil {
		return WrapDBError(fmt.Sprintf("failed to create %s table", migrationTableName), err)
	}
	return nil
}

func (tm *pgTableManager) getMostRecentSuccessVersion(ctx context.Context) (semver.Version, error) {
	query := fmt.Sprintf(`SELECT version FROM "%s".%s WHERE status = '%s' ORDER BY created DESC LIMIT 1`,
		tm.schemaName, migrationTableName, statusSuccess)
	var version semver.Version
	err := tm.connPool.QueryRow(ctx, query).Scan(&version)
	if err != nil {
		if goErrors.Is(err, pgx.ErrNoRows) {
			return semver.Version{}, nil // No rows found
		}
		return semver.Version{}, fmt.Errorf("failed to query most recent created version: %w", err)
	}
	return version, nil
}
