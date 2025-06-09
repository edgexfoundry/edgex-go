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
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// insertReadingCols defines the reading table columns in slice used in inserting readings
	insertReadingCols = []string{eventIdFKCol, deviceInfoIdFKCol, originCol, valueCol, binaryValueCol, objectValueCol}
)

func (c *Client) ReadingTotalCount() (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountReading())
}

func (c *Client) AllReadings(offset int, limit int) ([]model.Reading, errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	// query reading by origin descending with offset & limit
	readings, err := queryReadings(ctx, c.ConnPool, sqlQueryAllReadingWithPaginationDescByCol(originCol), offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to query all readings", err)
	}

	return readings, nil
}

// ReadingsByResourceName query readings by offset, limit and resource name
func (c *Client) ReadingsByResourceName(offset int, limit int, resourceName string) ([]model.Reading, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	// query reading by the resourceName and origin descending
	sqlStatement := sqlQueryAllReadingAndDescWithCondsAndPag(originCol, resourceNameCol)

	readings, err := queryReadings(context.Background(), c.ConnPool, sqlStatement, resourceName, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to query readings by resource '%s'", resourceName), err)
	}
	return readings, nil
}

// ReadingsByDeviceName query readings by offset, limit and device name
func (c *Client) ReadingsByDeviceName(offset int, limit int, name string) ([]model.Reading, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	// query reading by the deviceName and origin descending
	sqlStatement := sqlQueryAllReadingAndDescWithCondsAndPag(originCol, deviceNameCol)

	readings, err := queryReadings(context.Background(), c.ConnPool, sqlStatement, name, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to query readings by device '%s'", name), err)
	}
	return readings, nil
}

// ReadingsByDeviceNameAndResourceName query readings by offset, limit, device name and resource name
func (c *Client) ReadingsByDeviceNameAndResourceName(deviceName string, resourceName string, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	// query reading by the deviceName/resourceName and origin descending
	sqlStatement := sqlQueryAllReadingAndDescWithCondsAndPag(originCol, deviceNameCol, resourceNameCol)

	readings, err := queryReadings(context.Background(), c.ConnPool, sqlStatement, deviceName, resourceName, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to query readings by device '%s' and resource '%s'", deviceName, resourceName), err)
	}
	return readings, nil
}

// ReadingsByTimeRange query readings by origin within the time range with offset and limit
func (c *Client) ReadingsByTimeRange(start int64, end int64, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	sqlStatement := sqlQueryAllReadingWithPaginationAndTimeRangeDescByCol(originCol, originCol, nil)

	readings, err := queryReadings(ctx, c.ConnPool, sqlStatement, start, end, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, nil
}

// ReadingsByDeviceNameAndTimeRange query readings by the specified device, origin within the time range, offset, and limit
func (c *Client) ReadingsByDeviceNameAndTimeRange(deviceName string, start int64, end int64, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	sqlStatement := sqlQueryAllReadingWithPaginationAndTimeRangeDescByCol(originCol, originCol, nil, deviceNameCol)

	readings, err := queryReadings(ctx, c.ConnPool, sqlStatement, start, end, deviceName, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, nil
}

// ReadingsByResourceNameAndTimeRange query readings by the specified resource, origin within the time range, offset, and limit
func (c *Client) ReadingsByResourceNameAndTimeRange(resourceName string, start int64, end int64, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	sqlStatement := sqlQueryAllReadingWithPaginationAndTimeRangeDescByCol(originCol, originCol, nil, resourceNameCol)

	readings, err := queryReadings(ctx, c.ConnPool, sqlStatement, start, end, resourceName, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, nil
}

// ReadingsByDeviceNameAndResourceNameAndTimeRange query readings by the specified device and resource, origin within the time range, offset, and limit
func (c *Client) ReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, start int64, end int64, offset int, limit int) ([]model.Reading, errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	sqlStatement := sqlQueryAllReadingWithPaginationAndTimeRangeDescByCol(originCol, originCol, nil, deviceNameCol, resourceNameCol)

	readings, err := queryReadings(ctx, c.ConnPool, sqlStatement, start, end, deviceName, resourceName, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return readings, nil
}

// ReadingsByDeviceNameAndResourceNamesAndTimeRange query readings by the specified device and resourceName slice, origin within the time range, offset and limit
func (c *Client) ReadingsByDeviceNameAndResourceNamesAndTimeRange(deviceName string, resourceNames []string, start int64, end int64, offset, limit int) ([]model.Reading, errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	sqlStatement := sqlQueryAllReadingWithPaginationAndTimeRangeDescByCol(originCol, originCol,
		[]string{resourceNameCol}, deviceNameCol, resourceNameCol)

	// build the query args for the where condition using in querying readings
	queryArgs := []any{start, end, deviceName, resourceNames, offset, validLimit}
	// query readings
	readings, err := queryReadings(ctx, c.ConnPool, sqlStatement, queryArgs...)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	return readings, nil
}

// ReadingCountByDeviceName returns the count of Readings associated a specific Device from db
func (c *Client) ReadingCountByDeviceName(deviceName string) (uint32, errors.EdgeX) {
	sqlStatement := sqlQueryCountReadingByCol(deviceNameCol)

	return getTotalRowsCount(context.Background(), c.ConnPool, sqlStatement, deviceName)
}

// ReadingCountByResourceName returns the count of Readings associated a specific resource from db
func (c *Client) ReadingCountByResourceName(resourceName string) (uint32, errors.EdgeX) {
	sqlStatement := sqlQueryCountReadingByCol(resourceNameCol)

	return getTotalRowsCount(context.Background(), c.ConnPool, sqlStatement, resourceName)
}

// ReadingCountByDeviceNameAndResourceName returns the count of readings associated a specific device and resource values from db
func (c *Client) ReadingCountByDeviceNameAndResourceName(deviceName string, resourceName string) (uint32, errors.EdgeX) {
	sqlStatement := sqlQueryCountReadingByCol(deviceNameCol, resourceNameCol)

	return getTotalRowsCount(context.Background(), c.ConnPool, sqlStatement, deviceName, resourceName)
}

// ReadingCountByTimeRange returns the count of reading by origin within the time range from db
func (c *Client) ReadingCountByTimeRange(start int64, end int64) (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountReadingByTimeRangeCol(originCol, nil), start, end)
}

// ReadingCountByDeviceNameAndTimeRange returns the count of readings by origin within the time range and the specified device from db
func (c *Client) ReadingCountByDeviceNameAndTimeRange(deviceName string, start int64, end int64) (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountReadingByTimeRangeCol(originCol, nil, deviceNameCol), start, end, deviceName)
}

// ReadingCountByResourceNameAndTimeRange returns the count of readings by origin within the time range and the specified resource from db
func (c *Client) ReadingCountByResourceNameAndTimeRange(resourceName string, start int64, end int64) (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountReadingByTimeRangeCol(originCol, nil, resourceNameCol), start, end, resourceName)
}

// ReadingCountByDeviceNameAndResourceNameAndTimeRange returns the count of readings by origin within the time range
// associated with the specified device and resource from db
func (c *Client) ReadingCountByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, start int64, end int64) (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(),
		c.ConnPool,
		sqlQueryCountReadingByTimeRangeCol(originCol, nil, deviceNameCol, resourceNameCol),
		start,
		end,
		deviceName,
		resourceName)
}

// ReadingCountByDeviceNameAndResourceNamesAndTimeRange returns the count of readings by origin within the time range
// associated with the specified device and resourceName slice from db
func (c *Client) ReadingCountByDeviceNameAndResourceNamesAndTimeRange(deviceName string, resourceNames []string, start int64, end int64) (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(),
		c.ConnPool,
		sqlQueryCountReadingByTimeRangeCol(originCol, []string{resourceNameCol}, deviceNameCol, resourceNameCol),
		start, end, deviceName, resourceNames)
}

func (c *Client) LatestReadingByOffset(offset uint32) (model.Reading, errors.EdgeX) {
	ctx := context.Background()

	readings, err := queryReadings(ctx, c.ConnPool, sqlQueryAllReadingWithPaginationDescByCol(originCol), offset, 1)
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
			// reading type is NullReading
			nullReading := model.NullReading{
				BaseReading: baseReading,
				Value:       nil,
			}
			reading = nullReading
		}

		return reading, nil
	})

	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	return readings, nil
}

// deleteReadingsByOriginAndEventId delete the data rows with given sql statement and passed args
func deleteReadingsByOriginAndEventId(ctx context.Context, tx pgx.Tx, origin int64, eventId string) errors.EdgeX {
	sqlStatement := sqlDeleteTimeRangeByColumn(readingTableName, originCol, eventIdFKCol)
	commandTag, err := tx.Exec(
		ctx,
		sqlStatement,
		origin, eventId,
	)
	if commandTag.RowsAffected() == 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "no reading found", nil)
	}
	if err != nil {
		return pgClient.WrapDBError("reading(s) delete failed", err)
	}
	return nil
}

// deleteReadingsBySubQuery delete the readings with event_id in the range of the sub query
func deleteReadingsBySubQuery(ctx context.Context, tx pgx.Tx, subQuerySql string, args ...any) errors.EdgeX {
	sqlStatement := sqlDeleteByColumns(readingTableName, eventIdFKCol)
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
func (c *Client) addReadingsInTx(tx pgx.Tx, readings []model.Reading, eventId string) error {
	var readingDBModels []dbModels.Reading

	for _, r := range readings {
		baseReading := r.GetBaseReading()
		if baseReading.Id == "" {
			baseReading.Id = uuid.New().String()
		} else {
			_, err := uuid.Parse(baseReading.Id)
			if err != nil {
				return errors.NewCommonEdgeX(errors.KindInvalidId, "uuid parsing failed", err)
			}
		}

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
		case model.NullReading:
			readingDBModel = dbModels.Reading{
				BaseReading: baseReading,
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
			var objectValueBytes []byte
			var err error

			r := readingDBModels[i]
			if r.ObjectValue != nil {
				objectValueBytes, err = json.Marshal(r.ObjectValue)
				if err != nil {
					return nil, errors.NewCommonEdgeX(errors.KindServerError, "unable to JSON marshal reading ObjectValue", err)
				}
			}

			deviceInfoId, err := c.deviceInfoIdByReading(r)
			if err != nil {
				return nil, errors.NewCommonEdgeX(errors.KindServerError, "unable to retrieve deviceInfo", err)
			}

			return []any{
				eventId,
				deviceInfoId,
				r.Origin,
				r.Value,
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
