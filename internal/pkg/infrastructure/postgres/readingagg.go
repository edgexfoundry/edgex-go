//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"fmt"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
	dbModels "github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/postgres/models"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AllReadingsAggregation queries aggregated reading values using the specified SQL aggregation function.
func (c *Client) AllReadingsAggregation(aggregateFunc string, offset, limit int) ([]models.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithCondsAndPag(aggregateFunc, false)
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	readings, err := queryAggReadings(
		context.Background(),
		c.ConnPool,
		sqlStatement,
		pgx.NamedArgs{offsetCondition: offset, limitCondition: validLimit},
		aggregateFunc,
	)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to get the aggregated readings", err)
	}
	return readings, nil
}

// AllReadingsAggregationByTimeRange queries aggregated reading values within the given time range using the specified SQL aggregation function.
func (c *Client) AllReadingsAggregationByTimeRange(aggregateFun string, start, end int64, offset, limit int) ([]models.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithCondsAndPag(aggregateFun, true)
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	readings, err := queryAggReadings(
		context.Background(),
		c.ConnPool,
		sqlStatement,
		pgx.NamedArgs{startTimeCondition: start, endTimeCondition: end, offsetCondition: offset, limitCondition: validLimit},
		aggregateFun,
	)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to get the aggregated readings within the time range - start: %d, end: %d", start, end), err)
	}
	return readings, nil
}

// ReadingsAggregationByResourceName queries aggregated reading values by resource name using the specified SQL aggregation function.
func (c *Client) ReadingsAggregationByResourceName(resourceName string, aggregateFunc string, offset int, limit int) ([]models.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithCondsAndPag(aggregateFunc, false, resourceNameCol)
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	readings, err := queryAggReadings(
		context.Background(),
		c.ConnPool,
		sqlStatement,
		pgx.NamedArgs{resourceNameCol: resourceName, offsetCondition: offset, limitCondition: validLimit},
		aggregateFunc,
	)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to query readings by resource name '%s'", resourceName), err)
	}
	return readings, nil
}

// ReadingsAggregationByResourceNameAndTimeRange queries aggregated reading values by resource name within the given time range using the specified SQL aggregation function.
func (c *Client) ReadingsAggregationByResourceNameAndTimeRange(resourceName string, aggregateFun string, start int64, end int64, offset int, limit int) ([]models.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithCondsAndPag(aggregateFun, true, resourceNameCol)
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	readings, err := queryAggReadings(context.Background(),
		c.ConnPool,
		sqlStatement,
		pgx.NamedArgs{resourceNameCol: resourceName, startTimeCondition: start, endTimeCondition: end, offsetCondition: offset, limitCondition: validLimit},
		aggregateFun,
	)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to query readings by resource name '%s'", resourceName), err)
	}
	return readings, nil
}

// ReadingsAggregationByDeviceName queries aggregated reading values by device name using the specified SQL aggregation function.
func (c *Client) ReadingsAggregationByDeviceName(deviceName string, aggregateFunc string, offset int, limit int) ([]models.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithCondsAndPag(aggregateFunc, false, deviceNameCol)
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	readings, err := queryAggReadings(
		context.Background(),
		c.ConnPool,
		sqlStatement,
		pgx.NamedArgs{deviceNameCol: deviceName, offsetCondition: offset, limitCondition: validLimit},
		aggregateFunc,
	)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to query readings by device name '%s'", deviceName), err)
	}
	return readings, nil
}

// ReadingsAggregationByDeviceNameAndTimeRange queries aggregated reading values by device name within the given time range using the specified SQL aggregation function.
func (c *Client) ReadingsAggregationByDeviceNameAndTimeRange(deviceName string, aggregateFun string, start int64, end int64, offset int, limit int) ([]models.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithCondsAndPag(aggregateFun, true, deviceNameCol)
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	readings, err := queryAggReadings(context.Background(),
		c.ConnPool,
		sqlStatement,
		pgx.NamedArgs{deviceNameCol: deviceName, startTimeCondition: start, endTimeCondition: end, offsetCondition: offset, limitCondition: validLimit},
		aggregateFun,
	)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to query readings by device name '%s'", deviceName), err)
	}
	return readings, nil
}

// ReadingsAggregationByDeviceNameAndResourceName queries aggregated reading values by device name & resource name using the specified SQL aggregation function.
func (c *Client) ReadingsAggregationByDeviceNameAndResourceName(deviceName string, resourceName string, aggregateFunc string, offset int, limit int) ([]models.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithCondsAndPag(aggregateFunc, false, deviceNameCol, resourceNameCol)
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	readings, err := queryAggReadings(context.Background(),
		c.ConnPool,
		sqlStatement,
		pgx.NamedArgs{deviceNameCol: deviceName, resourceNameCol: resourceName, offsetCondition: offset, limitCondition: validLimit},
		aggregateFunc,
	)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to query readings by device '%s' and resource '%s'", deviceName, resourceName), err)
	}
	return readings, nil
}

// ReadingsAggregationByDeviceNameAndResourceNameAndTimeRange queries aggregated reading values by device name & resource name within the given time range using the specified SQL aggregation function.
func (c *Client) ReadingsAggregationByDeviceNameAndResourceNameAndTimeRange(deviceName string, resourceName string, aggregateFunc string, start int64, end int64, offset int, limit int) ([]models.Reading, errors.EdgeX) {
	sqlStatement := sqlQueryAggregateReadingWithCondsAndPag(aggregateFunc, true, deviceNameCol, resourceNameCol)
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	readings, err := queryAggReadings(
		context.Background(),
		c.ConnPool,
		sqlStatement,
		pgx.NamedArgs{deviceNameCol: deviceName, resourceNameCol: resourceName, startTimeCondition: start, endTimeCondition: end, offsetCondition: offset, limitCondition: validLimit},
		aggregateFunc,
	)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError,
			fmt.Sprintf("failed to query readings by device '%s' and resource '%s'", deviceName, resourceName), err)
	}
	return readings, nil
}

// queryAggReadings queries the aggregate readings data rows with given sql statement and passed args,
// converts the rows to map and unmarshal the data rows to the Reading model slice
func queryAggReadings(ctx context.Context, connPool *pgxpool.Pool, sql string, args pgx.NamedArgs, aggregateFunc string) ([]models.Reading, errors.EdgeX) {
	rows, err := connPool.Query(ctx, sql, args)
	if err != nil {
		return nil, pgClient.WrapDBError("query failed", err)
	}

	readings, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Reading, error) {
		readingDBModel, err := pgx.RowToStructByNameLax[dbModels.Reading](row)
		if err != nil {
			return nil, pgClient.WrapDBError("failed to convert row to map", err)
		}

		// convert the BaseReading fields to BaseReading struct defined in contract
		baseReading := readingDBModel.GetBaseReading()

		reading := models.NumericReading{
			BaseReading: baseReading,
		}
		if readingDBModel.NumericValue != nil {
			// reading type is NumericReading
			err = parseAggNumericReading(&reading, aggregateFunc, readingDBModel.NumericValue)
			if err != nil {
				return nil, pgClient.WrapDBError("read numeric value", err)
			}
		}

		return reading, nil
	})

	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	return readings, nil
}

// parseAggNumericReading converts the aggregate reading from numeric data type in PostgreSQL to the data type in Go
// based on the reading valueType and the quried aggregate function
func parseAggNumericReading(reading *models.NumericReading, aggregateFunc string, numericValue *pgtype.Numeric) errors.EdgeX {
	valueType := reading.ValueType

	switch aggregateFunc {
	case common.CountFunc:
		reading.ValueType = common.ValueTypeUint64
		// Always convert to Uint64 for COUNT function
		reading.NumericValue = numericValue.Int.Uint64()
	case common.AvgFunc:
		// Always convert to Float64 for AVG function
		val, err := numericValue.Float64Value()
		if err != nil {
			return pgClient.WrapDBError("failed to parse numeric data type to float64", err)
		}
		reading.ValueType = common.ValueTypeFloat64
		reading.NumericValue = val.Float64
	case common.SumFunc, common.MinFunc, common.MaxFunc:
		switch valueType {
		case common.ValueTypeFloat64, common.ValueTypeFloat32:
			// Convert to Float64 for float types
			val, err := numericValue.Float64Value()
			if err != nil {
				return pgClient.WrapDBError("read the float value", err)
			}
			reading.ValueType = common.ValueTypeFloat64
			reading.NumericValue = val.Float64
		case common.ValueTypeInt8, common.ValueTypeInt16, common.ValueTypeInt32, common.ValueTypeInt64:
			reading.ValueType = common.ValueTypeInt64
			// Convert to Int64 for int types
			reading.NumericValue = numericValue.Int.Int64()
		case common.ValueTypeUint8, common.ValueTypeUint16, common.ValueTypeUint32, common.ValueTypeUint64:
			reading.ValueType = common.ValueTypeUint64
			// Convert to Uint64 for uint types
			reading.NumericValue = numericValue.Int.Uint64()
		default:
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("unsupported valueType '%s' for aggregate function '%s'", valueType, aggregateFunc), nil)
		}
	default:
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("unexpected aggregateFunc type '%s", aggregateFunc), nil)
	}
	return nil
}
