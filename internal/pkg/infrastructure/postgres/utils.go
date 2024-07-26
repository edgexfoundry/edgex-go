package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
)

// getValidOffsetAndLimit returns the valid or default limit and offset from the given parameters
func getValidLimitAndOffset(offset, limit int) (int, int) {
	defaultOffset := 0
	defaultLimit := -1 //-1 limit means that clients want to retrieve all remaining records after offset from DB, so specifying -1 for end
	if offset < 0 {
		offset = defaultOffset
	}
	if limit < -1 {
		limit = defaultLimit
	}
	return offset, limit
}

// getValidStartAndEnd returns the valid start and end from the given parameters
func getValidStartAndEnd(start, end int64) (int64, int64, errors.EdgeX) {
	if end < start {
		return 0, 0, errors.NewCommonEdgeX(errors.KindContractInvalid, "end must be greater than start", nil)
	}
	return start, end, nil
}

// getValidRangeParameters returns the valid start, end, offset and limit from the given parameters
func getValidRangeParameters(start, end int64, offset, limit int) (int64, int64, int, int, errors.EdgeX) {
	var err errors.EdgeX
	end, start, err = getValidStartAndEnd(start, end)
	if err != nil {
		return 0, 0, 0, 0, errors.NewCommonEdgeXWrapper(err)
	}
	offset, limit = getValidLimitAndOffset(offset, limit)
	return start, end, offset, limit, nil
}

// getTotalRowsCount returns the total rows count from the given sql query
// Note: the sql query must be a count query, e.g. SELECT COUNT(*) FROM table_name
func getTotalRowsCount(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) (uint32, errors.EdgeX) {
	var rowCount int
	err := connPool.QueryRow(ctx, sql, args...).Scan(&rowCount)
	if err != nil {
		return 0, pgClient.WrapDBError("failed to query total rows count", err)
	}

	return uint32(rowCount), nil
}
