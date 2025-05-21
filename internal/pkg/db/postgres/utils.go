//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"embed"
	goErrors "errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/blang/semver/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

type versionDirNames []semver.Version

func (m versionDirNames) Len() int {
	return len(m)
}

func (m versionDirNames) Less(i, j int) bool {
	return m[i].LT(m[j])
}

func (m versionDirNames) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func getSortedSqlFileNames(embedFiles embed.FS, sqlFilesDir string) ([]string, error) {
	entries, err := embedFiles.ReadDir(sqlFilesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", sqlFilesDir, err)
	}

	sqlFiles := sqlFileNames{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()

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
		sqlFiles = append(sqlFiles, sqlFileName{order, toEmbedPath(sqlFilesDir, fileName)})
	}

	return sqlFiles.getSortedNames(), nil
}

func getSortedVersionDirNames(embedFiles embed.FS, sqlFilesDir string) ([]semver.Version, error) {
	entries, err := embedFiles.ReadDir(sqlFilesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", sqlFilesDir, err)
	}

	versionDirs := versionDirNames{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()

		// ignore folder whose name is not in the format of semver.Version
		version, parseErr := semver.Parse(dirName)
		// ignore mal-format version folder
		if parseErr != nil {
			continue
		}
		versionDirs = append(versionDirs, version)
	}

	sort.Sort(versionDirs)

	return versionDirs, nil
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

// toEmbedPath converts a path to the format required for go embed by using '/' as the separator.
func toEmbedPath(baseDir, fileName string) string {
	return filepath.ToSlash(filepath.Join(baseDir, fileName))
}
