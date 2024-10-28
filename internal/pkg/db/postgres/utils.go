//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	goErrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

var sqlFileNameRegexp = regexp.MustCompile(`([[:digit:]]+)-[[:word:]]+.sql`)

type sqlFileName struct {
	order int
	name  string
}

type sqlFileNames []sqlFileName

func (sf sqlFileNames) Len() int {
	return len(sf)
}

func (sf sqlFileNames) Less(i, j int) bool {
	return sf[i].order < sf[j].order
}

func (sf sqlFileNames) Swap(i, j int) {
	sf[i], sf[j] = sf[j], sf[i]
}

func (sf sqlFileNames) getSortedNames() []string {
	sort.Sort(sf)
	result := make([]string, len(sf))
	for i, f := range sf {
		result[i] = f.name
	}
	return result
}

func executeDBScripts(ctx context.Context, connPool *pgxpool.Pool, scriptsPath string) errors.EdgeX {
	if len(scriptsPath) == 0 {
		// skip script execution when the path is empty
		return nil
	}

	// get the sorted sql files
	sqlFiles, edgeXerr := sortedSqlFileNames(scriptsPath)
	if edgeXerr != nil {
		return edgeXerr
	}

	tx, err := connPool.Begin(ctx)
	if err != nil {
		return WrapDBError("failed to begin a transaction to execute sql files", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// execute sql files in the sequence of ordering prefix as a transaction
	for _, sqlFile := range sqlFiles {
		// read sql file content
		sqlContent, err := os.ReadFile(sqlFile)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to read sql file %s", sqlFile), err)
		}
		_, err = tx.Exec(ctx, string(sqlContent))
		if err != nil {
			return WrapDBError(fmt.Sprintf("failed to execute sql file %s", sqlFile), err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return WrapDBError("failed to commit transaction for executing sql files", err)
	}
	return nil
}

func sortedSqlFileNames(sqlFilesDir string) ([]string, errors.EdgeX) {
	sqlDir, err := os.Open(sqlFilesDir)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to open directory at %s", sqlFilesDir), err)
	}
	fileInfos, err := sqlDir.Readdir(0)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to read sql files from %s", sqlFilesDir), err)
	}

	sqlFiles := sqlFileNames{}
	for _, file := range fileInfos {
		// ignore directories
		if file.IsDir() {
			continue
		}

		fileName := file.Name()

		// ignore files whose name is not in the format of %d-%s.sql
		if !sqlFileNameRegexp.MatchString(fileName) {
			continue
		}

		var order int
		var fileNameWithoutOrder string
		_, err = fmt.Sscanf(fileName, "%d-%s", &order, &fileNameWithoutOrder)
		// ignore mal-format sql files
		if err != nil {
			continue
		}
		sqlFiles = append(sqlFiles, sqlFileName{order, filepath.Join(sqlFilesDir, fileName)})
	}
	return sqlFiles.getSortedNames(), nil
}

// When DB operation fails with error, the pgx DB library put a much detailed information in the pgxError.Detail.
func WrapDBError(message string, err error) errors.EdgeX {
	var pgErr *pgconn.PgError
	if goErrors.As(err, &pgErr) {
		if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			var errMsg string
			if message != "" {
				errMsg = message + ": "
			}
			errMsg += pgErr.Detail

			return errors.NewCommonEdgeX(errors.KindDuplicateName, errMsg, nil)
		}
		return errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("%s: %s %s", message, pgErr.Error(), pgErr.Detail), nil)
	} else if goErrors.Is(err, pgx.ErrNoRows) {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, message, err)
	}
	return errors.NewCommonEdgeX(errors.KindDatabaseError, message, err)
}
