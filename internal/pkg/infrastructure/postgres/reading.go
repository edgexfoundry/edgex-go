//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
	dbModels "github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/postgres/models"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// insertReadingCols defines the reading table columns in slice used in inserting readings
	insertReadingCols = []string{idCol, eventIdFKCol, deviceNameCol, profileNameCol, resourceNameCol, originCol, valueTypeCol, unitsCol, tagsCol, valueCol, mediaTypeCol, binaryValueCol, objectValueCol}
	// queryReadingCols defines the reading table columns in slice used in querying reading
	queryReadingCols = []string{idCol, deviceNameCol, profileNameCol, resourceNameCol, originCol, valueTypeCol, unitsCol, tagsCol, valueCol, objectValueCol, mediaTypeCol, binaryValueCol}
)

func (c *Client) ReadingTotalCount() (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCount(readingTableName))
}

func (c *Client) AllReadings(offset int, limit int) ([]model.Reading, errors.EdgeX) {
	ctx := context.Background()

	readings, err := queryReadings(ctx, c.ConnPool, sqlQueryAllWithPagination(readingTableName), offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to query all readings", err)
	}

	return readings, nil
}

// ReadingsByResourceName query readings by offset, limit and resource name
func (c *Client) ReadingsByResourceName(offset int, limit int, resourceName string) ([]model.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAllByColWithPagination(readingTableName, resourceNameCol)

	readings, err := queryReadings(context.Background(), c.ConnPool, sqlStatement, resourceName, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to query readings by resource '%s'", resourceName), err)
	}
	return readings, nil
}

// ReadingsByDeviceName query readings by offset, limit and device name
func (c *Client) ReadingsByDeviceName(offset int, limit int, name string) ([]model.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAllByColWithPagination(readingTableName, deviceNameCol)

	readings, err := queryReadings(context.Background(), c.ConnPool, sqlStatement, name, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to query readings by device '%s'", name), err)
	}
	return readings, nil
}

// ReadingsByDeviceNameAndResourceName query readings by offset, limit, device name and resource name
func (c *Client) ReadingsByDeviceNameAndResourceName(deviceName string, resourceName string, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAllByColWithPagination(readingTableName, deviceNameCol, resourceNameCol)

	readings, err := queryReadings(context.Background(), c.ConnPool, sqlStatement, deviceName, resourceName, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to query readings by device '%s' and resource '%s'", deviceName, resourceName), err)
	}
	return readings, nil
}

// ReadingsByTimeRange query readings by origin within the time range with offset and limit
func (c *Client) ReadingsByTimeRange(start int, end int, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	ctx := context.Background()
	sqlStatement := sqlQueryAllWithPaginationAndTimeRangeDescByCol(readingTableName, originCol, originCol, nil)

	readings, err := queryReadings(ctx, c.ConnPool, sqlStatement, start, end, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, nil
}

// ReadingsByDeviceNameAndTimeRange query readings by the specified device, origin within the time range, offset, and limit
func (c *Client) ReadingsByDeviceNameAndTimeRange(deviceName string, start int, end int, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	ctx := context.Background()
	sqlStatement := sqlQueryAllWithPaginationAndTimeRangeDescByCol(readingTableName, originCol, originCol, nil, deviceNameCol)

	readings, err := queryReadings(ctx, c.ConnPool, sqlStatement, start, end, deviceName, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, nil
}

// ReadingsByResourceNameAndTimeRange query readings by the specified resource, origin within the time range, offset, and limit
func (c *Client) ReadingsByResourceNameAndTimeRange(resourceName string, start int, end int, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	ctx := context.Background()
	sqlStatement := sqlQueryAllWithPaginationAndTimeRangeDescByCol(readingTableName, originCol, originCol, nil, resourceNameCol)

	readings, err := queryReadings(ctx, c.ConnPool, sqlStatement, start, end, resourceName, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, nil
}

// ReadingsByDeviceNameAndResourceNameAndTimeRange query readings by the specified device and resource, origin within the time range, offset, and limit
func (c *Client) ReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, start int, end int, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	ctx := context.Background()
	sqlStatement := sqlQueryAllWithPaginationAndTimeRangeDescByCol(readingTableName, originCol, originCol, nil, deviceNameCol, resourceNameCol)

	readings, err := queryReadings(ctx, c.ConnPool, sqlStatement, start, end, deviceName, resourceName, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, nil
}

// ReadingsByDeviceNameAndResourceNamesAndTimeRange query readings by the specified device and resourceName slice, origin within the time range, offset and limit
func (c *Client) ReadingsByDeviceNameAndResourceNamesAndTimeRange(deviceName string, resourceNames []string, start, end, offset, limit int) ([]model.Reading, uint32, errors.EdgeX) {
	ctx := context.Background()

	sqlStatement := sqlQueryAllWithPaginationAndTimeRangeDescByCol(readingTableName, originCol, originCol,
		[]string{resourceNameCol}, deviceNameCol, resourceNameCol)

	// build the query args for the where condition using in querying readings and reading count
	queryArgs := []any{start, end, deviceName, resourceNames}
	// make a copy for query count args as we don't need offset and limit while querying total count
	queryCountArgs := append([]any{}, queryArgs...)
	// add offset and limit for query args
	queryArgs = append(queryArgs, offset, limit)

	// query readings
	readings, err := queryReadings(ctx, c.ConnPool, sqlStatement, queryArgs...)
	if err != nil {
		return nil, 0, errors.NewCommonEdgeXWrapper(err)
	}

	// get the total count of readings based on the condition column names and query count args
	totalCount, err := getTotalRowsCount(context.Background(),
		c.ConnPool,
		sqlQueryCountByTimeRangeCol(readingTableName, originCol, []string{resourceNameCol}, deviceNameCol, resourceNameCol),
		queryCountArgs...)
	if err != nil {
		return nil, 0, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, totalCount, nil
}

// ReadingCountByDeviceName returns the count of Readings associated a specific Device from db
func (c *Client) ReadingCountByDeviceName(deviceName string) (uint32, errors.EdgeX) {
	sqlStatement := sqlQueryCountByCol(readingTableName, deviceNameCol)

	return getTotalRowsCount(context.Background(), c.ConnPool, sqlStatement, deviceName)
}

// ReadingCountByResourceName returns the count of Readings associated a specific resource from db
func (c *Client) ReadingCountByResourceName(resourceName string) (uint32, errors.EdgeX) {
	sqlStatement := sqlQueryCountByCol(readingTableName, resourceNameCol)

	return getTotalRowsCount(context.Background(), c.ConnPool, sqlStatement, resourceName)
}

// ReadingCountByDeviceNameAndResourceName returns the count of readings associated a specific device and resource values from db
func (c *Client) ReadingCountByDeviceNameAndResourceName(deviceName string, resourceName string) (uint32, errors.EdgeX) {
	sqlStatement := sqlQueryCountByCol(readingTableName, deviceNameCol, resourceNameCol)

	return getTotalRowsCount(context.Background(), c.ConnPool, sqlStatement, deviceName, resourceName)
}

// ReadingCountByTimeRange returns the count of reading by origin within the time range from db
func (c *Client) ReadingCountByTimeRange(start int, end int) (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByTimeRangeCol(readingTableName, originCol, nil), start, end)
}

// ReadingCountByDeviceNameAndTimeRange returns the count of readings by origin within the time range and the specified device from db
func (c *Client) ReadingCountByDeviceNameAndTimeRange(deviceName string, start int, end int) (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByTimeRangeCol(readingTableName, originCol, nil, deviceNameCol), start, end, deviceName)
}

// ReadingCountByResourceNameAndTimeRange returns the count of readings by origin within the time range and the specified resource from db
func (c *Client) ReadingCountByResourceNameAndTimeRange(resourceName string, start int, end int) (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByTimeRangeCol(readingTableName, originCol, nil, resourceNameCol), start, end, resourceName)
}

// ReadingCountByDeviceNameAndResourceNameAndTimeRange returns the count of readings by origin within the time range
// associated with the specified device and resource from db
func (c *Client) ReadingCountByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, start int, end int) (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(),
		c.ConnPool,
		sqlQueryCountByTimeRangeCol(readingTableName, originCol, nil, deviceNameCol, resourceNameCol),
		start,
		end,
		deviceName,
		resourceName)
}

func (c *Client) LatestReadingByOffset(offset uint32) (model.Reading, errors.EdgeX) {
	ctx := context.Background()

	readings, err := queryReadings(ctx, c.ConnPool, sqlQueryAllWithPaginationDescByCol(readingTableName, originCol), offset, 1)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to query all readings", err)
	}

	if len(readings) == 0 {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("no reading found with offset '%d'", offset), err)
	}
	return readings[0], nil
}

// queryReadings queries the data rows with given sql statement and passed args, converts the rows to map and unmarshal the data rows to the Reading model slice
func queryReadings(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) ([]model.Reading, errors.EdgeX) {
	rows, err := connPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgClient.WrapDBError("query failed", err)
	}

	readings, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Reading, error) {
		var reading model.Reading

		readingDBModel, err := pgx.RowToStructByNameLax[dbModels.Reading](row)
		if err != nil {
			return nil, pgClient.WrapDBError("failed to convert row to map", err)
		}

		// convert the BaseReading fields to BaseReading struct defined in contract
		baseReading := readingDBModel.GetBaseReading()

		if readingDBModel.BinaryValue != nil {
			// reading type is BinaryReading
			binaryReading := model.BinaryReading{
				BaseReading: baseReading,
				MediaType:   *readingDBModel.MediaType,
				BinaryValue: readingDBModel.BinaryValue,
			}
			reading = binaryReading
		} else if readingDBModel.ObjectValue != nil {
			// reading type is ObjectReading
			objReading := model.ObjectReading{
				BaseReading: baseReading,
				ObjectValue: readingDBModel.ObjectValue,
			}
			reading = objReading
		} else if readingDBModel.Value != nil {
			// reading type is SimpleReading
			simpleReading := model.SimpleReading{
				BaseReading: baseReading,
				Value:       *readingDBModel.Value,
			}
			reading = simpleReading
		} else {
			return reading, errors.NewCommonEdgeX(errors.KindServerError, "failed to convert reading to none of BinaryReading/ObjectReading/SimpleReading structs", nil)
		}

		return reading, nil
	})

	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	return readings, nil
}

// deleteReadings delete the data rows with given sql statement and passed args
func deleteReadings(ctx context.Context, tx pgx.Tx, args ...any) errors.EdgeX {
	sqlStatement := sqlDeleteByColumn(readingTableName, eventIdFKCol)
	commandTag, err := tx.Exec(
		ctx,
		sqlStatement,
		args...,
	)
	if commandTag.RowsAffected() == 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "no reading found", nil)
	}
	if err != nil {
		return pgClient.WrapDBError("reading(s) delete failed", err)
	}
	return nil
}

// deleteReadings delete the readings with event_id in the range of the sub query
func deleteReadingsBySubQuery(ctx context.Context, tx pgx.Tx, subQuerySql string, args ...any) errors.EdgeX {
	sqlStatement := sqlDeleteByColumn(readingTableName, eventIdFKCol)
	subQueryCond := "ANY ( " + subQuerySql + " )"
	sqlStatement = strings.Replace(sqlStatement, "$1", subQueryCond, 1)

	commandTag, err := tx.Exec(
		ctx,
		sqlStatement,
		args...,
	)
	if commandTag.RowsAffected() == 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "no reading found", nil)
	}
	if err != nil {
		return pgClient.WrapDBError("reading(s) delete failed", err)
	}
	return nil
}

// addReadingsInTx converts reading interface to BinaryReading/ObjectReading/SimpleReading structs first based on the reading value type
// and then perform the CopyFromSlice transaction to insert readings in batch
func addReadingsInTx(tx pgx.Tx, readings []model.Reading, eventId string) error {
	var readingDBModels []dbModels.Reading

	for _, r := range readings {
		baseReading := r.GetBaseReading()
		var readingDBModel dbModels.Reading

		switch contractReadingModel := r.(type) {
		case model.BinaryReading:
			// convert BinaryReading struct to Reading DB model
			readingDBModel = dbModels.Reading{
				BaseReading: baseReading,
				BinaryReading: dbModels.BinaryReading{
					BinaryValue: contractReadingModel.BinaryValue,
					MediaType:   &contractReadingModel.MediaType,
				},
			}
		case model.ObjectReading:
			// convert ObjectReading struct to Reading DB model
			readingDBModel = dbModels.Reading{
				BaseReading: baseReading,
				ObjectReading: dbModels.ObjectReading{
					ObjectValue: contractReadingModel.ObjectValue,
				},
			}
		case model.SimpleReading:
			// convert SimpleReading struct to Reading DB model
			readingDBModel = dbModels.Reading{
				BaseReading:   baseReading,
				SimpleReading: dbModels.SimpleReading{Value: &contractReadingModel.Value},
			}
		default:
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to convert reading to none of BinaryReading/ObjectReading/SimpleReading structs", nil)
		}
		readingDBModels = append(readingDBModels, readingDBModel)
	}

	// insert readingDBModels slice in batch
	_, err := tx.CopyFrom(
		context.Background(),
		strings.Split(readingTableName, "."),
		insertReadingCols,
		pgx.CopyFromSlice(len(readingDBModels), func(i int) ([]any, error) {
			var tagsBytes []byte
			var objectValueBytes []byte
			var err error

			r := readingDBModels[i]
			if r.Tags != nil {
				tagsBytes, err = json.Marshal(r.Tags)
				if err != nil {
					return nil, errors.NewCommonEdgeX(errors.KindServerError, "unable to JSON marshal reading tags", err)
				}
			}
			if r.ObjectValue != nil {
				objectValueBytes, err = json.Marshal(r.ObjectValue)
				if err != nil {
					return nil, errors.NewCommonEdgeX(errors.KindServerError, "unable to JSON marshal reading ObjectValue", err)
				}
			}
			return []any{
				r.Id,
				eventId,
				r.DeviceName,
				r.ProfileName,
				r.ResourceName,
				r.Origin,
				r.ValueType,
				r.Units,
				tagsBytes,
				r.Value,
				r.MediaType,
				r.BinaryValue,
				objectValueBytes,
			}, nil
		}),
	)
	if err != nil {
		return pgClient.WrapDBError("failed to insert readings in batch", err)
	}

	return nil
}
