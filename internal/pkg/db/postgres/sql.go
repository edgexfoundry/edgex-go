//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func executeSqlFilesDir(ctx context.Context, connPool *pgxpool.Pool, embedSqlFiles embed.FS, scriptsDir string) error {
	if len(scriptsDir) == 0 {
		// skip script execution when the path is empty
		return nil
	}

	// get the sorted sql files
	sqlFiles, err := getSortedSqlFileNames(embedSqlFiles, scriptsDir)
	if err != nil {
		return err
	}

	// begin a transaction to execute all sql files under the same directory
	tx, err := connPool.Begin(ctx)
	if err != nil {
		return WrapDBError("failed to begin a transaction to execute sql files", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// execute sql files in the sequence of ordering prefix as a transaction
	for _, sqlFile := range sqlFiles {
		if err = executeSqlFile(ctx, tx, embedSqlFiles, sqlFile); err != nil {
			return err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return WrapDBError("failed to commit transaction for executing sql files", err)
	}
	return nil
}

func executeSqlFile(ctx context.Context, tx pgx.Tx, embedSqlFileDir embed.FS, sqlFilePath string) error {
	// read sql file content
	sqlContent, err := embedSqlFileDir.ReadFile(sqlFilePath)
	if err != nil {
		return fmt.Errorf("failed to read sql file %s: %w", sqlFilePath, err)
	}
	sqlContentStr := string(sqlContent)
	_, err = tx.Exec(ctx, sqlContentStr)
	if err != nil {
		//special handling if we know the position of the error
		pgErr, ok := err.(*pgconn.PgError)
		if ok && pgErr.Position > 0 {
			strC := sqlContentStr
			code := strC[pgErr.Position-1:]
			return fmt.Errorf("failed to execute sql script %s with err %w code at error position:\n  %s", sqlFilePath, err, code)
		}
		return WrapDBError(fmt.Sprintf("failed to execute sql file %s", sqlFilePath), err)
	}
	return nil
}
