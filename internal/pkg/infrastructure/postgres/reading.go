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

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var insertBaseReadingCol = []string{idCol, eventIdFKCol, deviceNameCol, profileNameCol, resourceNameCol, originCol, valueTypeCol, unitsCol, tagsCol}

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
		var baseReading model.BaseReading

		readingMap, err := pgx.RowToMap(row)
		if err != nil {
			return nil, pgClient.WrapDBError("failed to convert row to map", err)
		}

		// retrieve the BaseReading fields from the map
		if id, ok := readingMap[idCol].([16]uint8); ok {
			baseReading.Id = fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:16])
		}
		if origin, ok := readingMap[originCol].(int64); ok {
			baseReading.Origin = origin
		}
		if deviceName, ok := readingMap[deviceNameCol].(string); ok {
			baseReading.DeviceName = deviceName
		}
		if profileName, ok := readingMap[profileNameCol].(string); ok {
			baseReading.ProfileName = profileName
		}
		if resourceName, ok := readingMap[resourceNameCol].(string); ok {
			baseReading.ResourceName = resourceName
		}
		if valueType, ok := readingMap[valueTypeCol].(string); ok {
			baseReading.ValueType = valueType
		}
		if units, ok := readingMap[unitsCol].(string); ok {
			baseReading.Units = units
		}
		if tags, ok := readingMap[tagsCol].(map[string]any); ok {
			baseReading.Tags = tags
		}

		value, ok := readingMap[valueCol].(string)
		if ok && value != "" {
			// reading type is SimpleReading
			simpleReading := model.SimpleReading{
				BaseReading: baseReading,
				Value:       value,
			}
			reading = simpleReading
		} else {
			// reading type is not SimpleReading, check if the reading belongs to either BinaryReading or ObjectReading
			if baseReading.ValueType == common.ValueTypeBinary {
				// reading type is BinaryReading
				binaryReading := model.BinaryReading{
					BaseReading: baseReading,
				}
				if mediaType, ok := readingMap[mediaTypeCol].(string); ok {
					binaryReading.MediaType = mediaType
				}
				if binaryValue, ok := readingMap[binaryValueCol].([]byte); ok {
					binaryReading.BinaryValue = binaryValue
				}
				reading = binaryReading
			} else if baseReading.ValueType == common.ValueTypeObject {
				// reading type is ObjectReading
				objReading := model.ObjectReading{
					BaseReading: baseReading,
				}
				if objValue, ok := readingMap[objectValueCol]; ok {
					objReading.ObjectValue = objValue
				}
				reading = objReading
			} else {
				return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("unknown reading value type '%s'", baseReading.ValueType), nil)
			}
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
	var binaryReadings []model.BinaryReading
	var objReadings []model.ObjectReading
	var simpleReadings []model.SimpleReading

	var err error

	for _, r := range readings {
		baseReading := r.GetBaseReading()

		valueType := baseReading.ValueType
		if valueType == common.ValueTypeBinary {
			// convert reading to BinaryReading struct
			b, ok := r.(model.BinaryReading)
			if !ok {
				return errors.NewCommonEdgeX(errors.KindServerError, "failed to convert reading to BinaryReading model", nil)
			}
			binaryReadings = append(binaryReadings, b)
		} else if valueType == common.ValueTypeObject {
			// convert reading to ObjectReading struct
			o, ok := r.(model.ObjectReading)
			if !ok {
				return errors.NewCommonEdgeX(errors.KindServerError, "failed to convert reading to ObjectReading model", nil)
			}
			objReadings = append(objReadings, o)
		} else {
			// convert reading to SimpleReading struct
			s, ok := r.(model.SimpleReading)
			if !ok {
				return errors.NewCommonEdgeX(errors.KindServerError, "failed to convert reading to SimpleReading model", nil)
			}
			simpleReadings = append(simpleReadings, s)
		}
	}

	// insert binary readings in batch
	if len(binaryReadings) > 0 {
		binaryReadingCols := append(insertBaseReadingCol, mediaTypeCol, binaryValueCol)

		_, err = tx.CopyFrom(
			context.Background(),
			strings.Split(readingTableName, "."),
			binaryReadingCols,
			pgx.CopyFromSlice(len(binaryReadings), func(i int) ([]any, error) {
				tagsBytes, err := json.Marshal(binaryReadings[i].Tags)
				if err != nil {
					return nil, errors.NewCommonEdgeX(errors.KindServerError, "unable to JSON marshal reading tags", err)
				}
				return []any{
					binaryReadings[i].Id,
					eventId,
					binaryReadings[i].DeviceName,
					binaryReadings[i].ProfileName,
					binaryReadings[i].ResourceName,
					binaryReadings[i].Origin,
					binaryReadings[i].ValueType,
					binaryReadings[i].Units,
					tagsBytes,
					binaryReadings[i].MediaType,
					binaryReadings[i].BinaryValue,
				}, nil
			}),
		)
	}

	// insert object readings in batch
	if len(objReadings) > 0 {
		objReadingCols := append(insertBaseReadingCol, objectValueCol)

		_, err = tx.CopyFrom(
			context.Background(),
			strings.Split(readingTableName, "."),
			objReadingCols,
			pgx.CopyFromSlice(len(objReadings), func(i int) ([]any, error) {
				tagsBytes, err := json.Marshal(objReadings[i].Tags)
				if err != nil {
					return nil, errors.NewCommonEdgeX(errors.KindServerError, "unable to JSON marshal reading tags", err)
				}
				return []any{
					objReadings[i].Id,
					eventId,
					objReadings[i].DeviceName,
					objReadings[i].ProfileName,
					objReadings[i].ResourceName,
					objReadings[i].Origin,
					objReadings[i].ValueType,
					objReadings[i].Units,
					tagsBytes,
					objReadings[i].ObjectValue,
				}, nil
			}),
		)
	}

	// insert simple readings in batch
	if len(simpleReadings) > 0 {
		simpleReadingCols := append(insertBaseReadingCol, valueCol)

		_, err = tx.CopyFrom(
			context.Background(),
			strings.Split(readingTableName, "."),
			simpleReadingCols,
			pgx.CopyFromSlice(len(simpleReadings), func(i int) ([]any, error) {
				tagsBytes, err := json.Marshal(simpleReadings[i].Tags)
				if err != nil {
					return nil, errors.NewCommonEdgeX(errors.KindServerError, "unable to JSON marshal reading tags", err)
				}
				return []any{
					simpleReadings[i].Id,
					eventId,
					simpleReadings[i].DeviceName,
					simpleReadings[i].ProfileName,
					simpleReadings[i].ResourceName,
					simpleReadings[i].Origin,
					simpleReadings[i].ValueType,
					simpleReadings[i].Units,
					tagsBytes,
					simpleReadings[i].Value,
				}, nil
			}),
		)
	}
	if err != nil {
		return pgClient.WrapDBError("failed to insert readings", err)
	}

	return nil
}
