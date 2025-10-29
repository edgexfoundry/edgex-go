//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
)

// getValidOffsetAndLimit returns the valid or default offset and limit from the given parameters
// Note: the returned limit can be an integer or nil (which means that clients want to retrieve all remaining rows after offset)
func getValidOffsetAndLimit(offset, limit int) (int, any) {
	defaultOffset := 0
	if offset < 0 {
		offset = defaultOffset
	}

	// Since PostgreSQL does not support negative limit, we need to set the default limit to nil,
	// nil limit means that clients want to retrieve all remaining rows after offset from the DB
	var defaultLimit any = nil
	if limit < 0 {
		return offset, defaultLimit
	} else {
		return offset, limit
	}
}

// getValidStartAndEnd returns the valid start and end from the given parameters
func getValidStartAndEnd(start, end int64) (int64, int64, errors.EdgeX) {
	if end < start {
		return 0, 0, errors.NewCommonEdgeX(errors.KindContractInvalid, "end must be greater than start", nil)
	}
	return start, end, nil
}

// getValidStartAndEndTime returns the valid start and end from the given parameters in time.Time type
func getValidStartAndEndTime(start, end int64) (time.Time, time.Time, errors.EdgeX) {
	start, end, err := getValidStartAndEnd(start, end)
	if err != nil {
		return time.Time{}, time.Time{}, errors.NewCommonEdgeXWrapper(err)
	}

	startTime, endTime := getUTCStartAndEndTime(start, end)
	return startTime, endTime, nil
}

// getValidRangeParameters returns the valid start, end, offset and limit from the given parameters for querying data from the PostgreSQL database
// Note: the returned limit can be an integer or nil (which means that clients want to retrieve all remaining rows after offset)
func getValidRangeParameters(start, end int64, offset, limit int) (int64, int64, int, any, errors.EdgeX) {
	start, end, err := getValidStartAndEnd(start, end)
	if err != nil {
		return 0, 0, 0, nil, errors.NewCommonEdgeXWrapper(err)
	}

	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	return start, end, offset, validLimit, nil
}

// getValidTimeRangeParameters returns the valid start, end, offset and limit from the given parameters for querying data from the PostgreSQL database
// Note: the returned start and end are in time.Time type
// Note: the returned limit can be an integer or nil (which means that clients want to retrieve all remaining rows after offset)
func getValidTimeRangeParameters(start, end int64, offset, limit int) (time.Time, time.Time, int, any, errors.EdgeX) {
	startTime, endTime, err := getValidStartAndEndTime(start, end)
	if err != nil {
		return time.Time{}, time.Time{}, 0, nil, errors.NewCommonEdgeXWrapper(err)
	}

	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	return startTime, endTime, offset, validLimit, nil
}

// getTotalRowsCount returns the total rows count from the given sql query
// Note: the sql query must be a count query, e.g. SELECT COUNT(*) FROM table_name
func getTotalRowsCount(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) (int64, errors.EdgeX) {
	var rowCount int64
	err := connPool.QueryRow(ctx, sql, args...).Scan(&rowCount)
	if err != nil {
		return 0, pgClient.WrapDBError("failed to query total rows count", err)
	}

	return rowCount, nil
}

// getUTCStartAndEndTime returns the UTC start and end time from the given start and end timestamp
func getUTCStartAndEndTime(start, end int64) (time.Time, time.Time) {
	return getUTCTime(start), getUTCTime(end)
}

func getUTCTime(timestamp int64) time.Time {
	return time.UnixMilli(timestamp).UTC()
}
