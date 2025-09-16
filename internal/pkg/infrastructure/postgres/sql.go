//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This file contains the common SQL statements for interacting with the PostgreSQL database.
// All the arguments need to manually pass to the actual jackc/pgx operation functions.

package postgres

import (
	"fmt"
	"slices"
	"strings"
)

const (
	eventColumns             = "event.id, devicename, profilename, sourcename, origin, tags"
	readingColumns           = "event_id, origin, value, numeric_value, binaryvalue, objectvalue, devicename, profilename, resourcename, valuetype, units, mediatype, tags"
	aggReadingColumn         = "numeric_value"
	aggReadingGroupByColumns = "devicename, profilename, resourcename, valuetype"
)

// ----------------------------------------------------------------------------------
// SQL statements for INSERT operations
// ----------------------------------------------------------------------------------

// sqlInsert returns the SQL statement for inserting a new row with the given columns into the table.
func sqlInsert(table string, columns ...string) string {
	columnCount := len(columns)
	columnNames := strings.Join(columns, ", ")
	valuePlaceholders := make([]string, columnCount)

	for i := 0; i < columnCount; i++ {
		valuePlaceholders[i] = fmt.Sprintf("$%d", i+1)
	}

	valueNames := strings.Join(valuePlaceholders, ", ")
	return fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s)", table, columnNames, valueNames)
}

// ----------------------------------------------------------------------------------
// SQL statements for SELECT operations
// ----------------------------------------------------------------------------------

// sqlQueryFieldsByCol returns the SQL statement for selecting the given fields of rows from the table by the conditions composed of given columns
func sqlQueryFieldsByCol(table string, fields []string, columns ...string) string {
	whereCondition := constructWhereCondition(columns...)
	queryFieldStr := strings.Join(fields, ", ")

	return fmt.Sprintf("SELECT %s FROM %s WHERE %s", queryFieldStr, table, whereCondition)
}

// sqlQueryEventIdFieldsByCol returns the SQL statement for selecting the event.id of rows from the event table by the conditions composed of given columns
func sqlQueryEventIdFieldsByCol(columns ...string) string {
	whereCondition := constructWhereNamedArgCondition(columns...)

	return fmt.Sprintf("SELECT event.id FROM %s JOIN %s on event.device_info_id = device_info.id WHERE %s", eventTableName, deviceInfoTableName, whereCondition)
}

// sqlQueryFieldsByColAndLikePat returns the SQL statement for selecting the given fields of rows from the table by the conditions composed of given columns with LIKE operator
func sqlQueryFieldsByColAndLikePat(table string, fields []string, columns ...string) string {
	whereCondition := constructWhereLikeCond(columns...)
	queryFieldStr := strings.Join(fields, ", ")

	return fmt.Sprintf("SELECT %s FROM %s WHERE %s", queryFieldStr, table, whereCondition)
}

// sqlQueryContentWithPaginationAsNamedArgs returns the SQL statement for selecting content column from the table with pagination
func sqlQueryAllWithPaginationAsNamedArgs(table string) string {
	return fmt.Sprintf("SELECT * FROM %s OFFSET @%s LIMIT @%s", table, offsetCondition, limitCondition)
}

// sqlQueryAllWithNamedArgConds returns the SQL statement for selecting all rows from the table by the given columns composed of the where condition
func sqlQueryAllWithNamedArgConds(table string, columns ...string) string {
	whereCondition := constructWhereNamedArgCondition(columns...)

	return fmt.Sprintf("SELECT * FROM %s WHERE %s", table, whereCondition)
}

// sqlQueryAllWithPaginationDescByCol returns the SQL statement for selecting all rows from the table with the pagination and desc by descCol
func sqlQueryAllWithPaginationDescByCol(table string, descCol string) string {
	return fmt.Sprintf("SELECT * FROM %s ORDER BY %s DESC OFFSET @%s LIMIT @%s", table, descCol, offsetCondition, limitCondition)
}

// sqlQueryAllEventWithPaginationDescByCol returns the SQL statement for selecting all rows from the event table with the pagination and desc by descCol
func sqlQueryAllEventWithPaginationDescByCol(descCol string) string {
	return fmt.Sprintf("SELECT %s FROM %s JOIN %s on event.device_info_id = device_info.id WHERE %s = false ORDER BY %s DESC OFFSET @%s LIMIT @%s", eventColumns, eventTableName, deviceInfoTableName, markDeletedCol, descCol, offsetCondition, limitCondition)
}

// sqlQueryAllReadingWithPaginationDescByCol returns the SQL statement for selecting all rows from the reading table with the pagination and desc by descCol
func sqlQueryAllReadingWithPaginationDescByCol(descCol string) string {
	return fmt.Sprintf("SELECT %s FROM %s JOIN %s on reading.device_info_id = device_info.id WHERE %s = false ORDER BY %s DESC OFFSET @%s LIMIT @%s", readingColumns, readingTableName, deviceInfoTableName, markDeletedCol, descCol, offsetCondition, limitCondition)
}

// sqlQueryAllReadingAndDescWithConds returns the SQL statement for selecting all rows from the table by the given columns composed of the where condition with descending by descCol
func sqlQueryAllReadingAndDescWithConds(descCol string, columns ...string) string {
	whereCondition := constructWhereNamedArgCondition(columns...)

	return fmt.Sprintf("SELECT %s FROM %s JOIN %s on reading.device_info_id = device_info.id WHERE %s = false AND %s ORDER BY %s DESC", readingColumns, readingTableName, deviceInfoTableName, markDeletedCol, whereCondition, descCol)
}

// sqlQueryAllEventAndDescWithCondsAndPage returns the SQL statement for selecting all rows from the event table by the given columns composed of the where condition
// with descending by descCol and pagination
func sqlQueryAllEventAndDescWithCondsAndPage(descCol string, columns ...string) string {
	whereCondition := constructWhereNamedArgCondition(columns...)

	return fmt.Sprintf(
		"SELECT %s FROM %s join %s on event.device_info_id = device_info.id WHERE %s = false AND %s ORDER BY %s DESC OFFSET @%s LIMIT @%s",
		eventColumns, eventTableName, deviceInfoTableName, markDeletedCol,
		whereCondition, descCol,
		offsetCondition, limitCondition)
}

// sqlQueryAllReadingAndDescWithCondsAndPag returns the SQL statement for selecting all rows from the reading table by the given columns composed of the where condition
// with descending by descCol and pagination
func sqlQueryAllReadingAndDescWithCondsAndPag(descCol string, columns ...string) string {
	whereCondition := constructWhereNamedArgCondition(columns...)

	return fmt.Sprintf(
		"SELECT %s FROM %s join %s on reading.device_info_id = device_info.id WHERE %s = false AND %s ORDER BY %s DESC OFFSET @%s LIMIT @%s",
		readingColumns, readingTableName, deviceInfoTableName, markDeletedCol,
		whereCondition, descCol,
		offsetCondition, limitCondition)
}

// sqlQueryAggregateReadingWithCondsAndPag returns the SQL statement for calculating the aggregated reading table by the given columns composed of the where condition
// If hasTimeRange is true, the time range will be defined in the where condition as well
// Results are grouped by deviceName, resourceName, profileName, and valueType, and ordered by deviceName in ascending order and pagination.
func sqlQueryAggregateReadingWithCondsAndPag(aggFunc string, hasTimeRange bool, columns ...string) string {
	statement := fmt.Sprintf(
		"SELECT %s(%s) AS numeric_value, %s FROM %s JOIN %s ON reading.device_info_id = device_info.id WHERE %s = false",
		aggFunc, aggReadingColumn, aggReadingGroupByColumns, readingTableName, deviceInfoTableName, markDeletedCol,
	)

	// Check whether to construct the where conditions
	if len(columns) > 0 {
		whereCondition := constructWhereNamedArgCondition(columns...)
		statement += fmt.Sprintf(" AND %s", whereCondition)
	}

	// Check whether to construct the time range conditions
	if hasTimeRange {
		statement += fmt.Sprintf(" AND %s", fmt.Sprintf("%s >= @%s", originCol, startTimeCondition))
		statement += fmt.Sprintf(" AND %s", fmt.Sprintf("%s <= @%s", originCol, endTimeCondition))
	}

	// Add the "group by" and "order by" conditions
	statement += fmt.Sprintf(" GROUP BY %s ORDER BY %s OFFSET @%s LIMIT @%s",
		aggReadingGroupByColumns, deviceNameCol, offsetCondition, limitCondition)
	return statement
}

// sqlQueryAllEventAndDescWithCondsAndPagAndUpperLimitTime returns the SQL statement for selecting all rows from the event table by the given columns composed of the where condition
// with descending by descCol and pagination
func sqlQueryAllEventAndDescWithCondsAndPagAndUpperLimitTime(descCol string, upperLimitTimeRangeCol string, columns ...string) string {
	whereCondition := constructWhereNamedArgCondWithTimeRange("", upperLimitTimeRangeCol, nil, columns...)

	return fmt.Sprintf(
		"SELECT %s FROM %s join %s on event.device_info_id = device_info.id WHERE %s = false AND %s ORDER BY %s DESC OFFSET @%s LIMIT @%s",
		eventColumns, eventTableName, deviceInfoTableName, markDeletedCol,
		whereCondition, descCol,
		offsetCondition, limitCondition)
}

// sqlQueryAllWithPaginationAndTimeRangeAsNamedArgs returns the SQL statement for selecting all rows from the table with pagination and a time range.
func sqlQueryAllWithPaginationAndTimeRangeAsNamedArgs(table string) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE %s >= @%s AND %s <= @%s ORDER BY %s OFFSET @%s LIMIT @%s",
		table, createdCol, startTimeCondition, createdCol, endTimeCondition, createdCol, offsetCondition, limitCondition)
}

// sqlQueryAllEventWithPaginationAndTimeRangeDescByCol returns the SQL statement for selecting all rows from the event table with the arrayColNames slice,
// provided columns with pagination and a time range by timeRangeCol, desc by descCol
func sqlQueryAllEventWithPaginationAndTimeRangeDescByCol(timeRangeCol string, descCol string) string {
	whereCondition := constructWhereNamedArgCondWithTimeRange(timeRangeCol, timeRangeCol, nil)

	return fmt.Sprintf(
		"SELECT %s FROM %s join %s on event.device_info_id = device_info.id WHERE %s = false AND %s ORDER BY %s DESC OFFSET @%s LIMIT @%s",
		eventColumns, eventTableName, deviceInfoTableName, markDeletedCol,
		whereCondition, descCol,
		offsetCondition, limitCondition)
}

// sqlQueryAllReadingWithPaginationAndTimeRangeDescByCol returns the SQL statement for selecting all rows from the reading table with the arrayColNames slice,
// provided columns with pagination and a time range by timeRangeCol, desc by descCol
func sqlQueryAllReadingWithPaginationAndTimeRangeDescByCol(timeRangeCol string, descCol string, arrayColNames []string, columns ...string) string {
	whereCondition := constructWhereNamedArgCondWithTimeRange(timeRangeCol, timeRangeCol, arrayColNames, columns...)

	return fmt.Sprintf(
		"SELECT %s FROM %s join %s on reading.device_info_id = device_info.id WHERE %s = false AND %s ORDER BY %s DESC OFFSET @%s LIMIT @%s",
		readingColumns, readingTableName, deviceInfoTableName, markDeletedCol,
		whereCondition, descCol,
		offsetCondition, limitCondition)
}

// sqlQueryAllByStatusWithPaginationAndTimeRange returns the SQL statement for selecting all rows from the table by status with pagination and a time range.
func sqlQueryAllByStatusWithPaginationAndTimeRange(table string) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE %s = $1 AND %s >= $2 AND %s <= $3 ORDER BY %s OFFSET $4 LIMIT $5", table, statusCol, createdCol, createdCol, createdCol)
}

// sqlQueryAllByColWithPaginationAndTimeRange returns the SQL statement for selecting all rows from the table by the given columns with pagination and a time range.
func sqlQueryAllByColWithPaginationAndTimeRange(table string, columns ...string) string {
	whereCondition := constructWhereNamedArgCondWithTimeRange(createdCol, createdCol, nil, columns...)

	return fmt.Sprintf(
		"SELECT * FROM %s WHERE %s ORDER BY %s OFFSET @%s LIMIT @%s",
		table, whereCondition, createdCol,
		offsetCondition, limitCondition)
}

// sqlQueryAllById returns the SQL statement for selecting all rows from the table by id.
func sqlQueryAllById(table string) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE %s = $1", table, idCol)
}

// sqlQueryAllEventById returns the SQL statement for selecting all rows from the event table by id.
func sqlQueryAllEventById() string {
	return fmt.Sprintf("SELECT %s FROM %s JOIN %s on event.device_info_id = device_info.id WHERE core_data.event.id=@%s", eventColumns, eventTableName, deviceInfoTableName, idCol)
}

// sqlQueryContentById returns the SQL statement for selecting content column by the specified id.
func sqlQueryContentById(table string) string {
	return fmt.Sprintf("SELECT content FROM %s WHERE %s = $1", table, idCol)
}

// sqlQueryContent returns the SQL statement for selecting content column in the table for all entries
func sqlQueryContent(table string) string {
	return fmt.Sprintf("SELECT content FROM %s", table)
}

// sqlQueryContentWithPagination returns the SQL statement for selecting content column from the table with pagination
func sqlQueryContentWithPagination(table string) string {
	return fmt.Sprintf("SELECT content FROM %s ORDER BY COALESCE((content->>'%s')::bigint, 0) OFFSET $1 LIMIT $2", table, createdField)
}

// sqlQueryContentWithPaginationAsNamedArgs returns the SQL statement for selecting content column from the table with pagination
func sqlQueryContentWithPaginationAsNamedArgs(table string) string {
	return fmt.Sprintf("SELECT content FROM %s ORDER BY COALESCE((content->>'%s')::bigint, 0) OFFSET @%s LIMIT @%s", table, createdField, offsetCondition, limitCondition)
}

// sqlQueryContentWithTimeRangeAndPaginationAsNamedArgs returns the SQL statement for selecting content column from the table by the given time range with pagination
func sqlQueryContentWithTimeRangeAndPaginationAsNamedArgs(table string) string {
	return fmt.Sprintf("SELECT content FROM %s WHERE COALESCE((content->>'%s')::bigint, 0) BETWEEN @%s AND @%s AND content @> @%s::jsonb ORDER BY COALESCE((content->>'%s')::bigint, 0) OFFSET @%s LIMIT @%s",
		table, createdField, startTimeCondition, endTimeCondition, jsonContentCondition, createdField, offsetCondition, limitCondition)
}

// sqlQueryContentByJSONField returns the SQL statement for selecting content column in the table by the given JSON query string
func sqlQueryContentByJSONField(table string) string {
	return fmt.Sprintf("SELECT content FROM %s WHERE content @> $1::jsonb", table)
}

// sqlQueryContentByJSONFieldWithPaginationAsNamedArgs returns the SQL statement for selecting content column in the table by the given JSON query string with pagination
func sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(table string) string {
	return fmt.Sprintf("SELECT content FROM %s WHERE content @> @%s::jsonb ORDER BY COALESCE((content->>'%s')::bigint, 0) OFFSET @%s LIMIT @%s",
		table, jsonContentCondition, createdField, offsetCondition, limitCondition)
}

// sqlCheckExistsById returns the SQL statement for checking if a row exists in the table by id.
func sqlCheckExistsById(table string) string {
	return fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = $1)", table, idCol)
}

// sqlCheckExistsByCol returns the SQL statement for checking if a row exists in the table by where condition.
func sqlCheckExistsByCol(table string, columns ...string) string {
	whereCondition := constructWhereCondition(columns...)
	return fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s)", table, whereCondition)
}

// sqlCheckExistsByJSONField returns the SQL statement for checking if a row exists by query the JSON field in content column.
func sqlCheckExistsByJSONField(table string) string {
	return fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE content @> $1::jsonb)", table)
}

// sqlQueryCount returns the SQL statement for counting the number of rows in the table.
func sqlQueryCount(table string) string {
	return fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
}

// sqlQueryCountEvent returns the SQL statement for counting the number of rows in the table.
func sqlQueryCountEvent() string {
	return fmt.Sprintf("SELECT COUNT(*) FROM %s join %s on event.device_info_id = device_info.id WHERE %s = false", eventTableName, deviceInfoTableName, markDeletedCol)
}

// sqlQueryCountReading returns the SQL statement for counting the number of rows in the table.
func sqlQueryCountReading() string {
	return fmt.Sprintf("SELECT COUNT(*) FROM %s join %s on reading.device_info_id = device_info.id WHERE %s = false", readingTableName, deviceInfoTableName, markDeletedCol)
}

// sqlQueryCountEventByCol returns the SQL statement for counting the number of rows in the table by the given column name.
func sqlQueryCountEventByCol(columns ...string) string {
	whereCondition := constructWhereNamedArgCondition(columns...)
	return fmt.Sprintf("SELECT COUNT(*) FROM %s join %s on event.device_info_id = device_info.id WHERE %s = false AND %s", eventTableName, deviceInfoTableName, markDeletedCol, whereCondition)
}

// sqlQueryCountReadingByCol returns the SQL statement for counting the number of rows in the table by the given column name.
func sqlQueryCountReadingByCol(columns ...string) string {
	whereCondition := constructWhereNamedArgCondition(columns...)
	return fmt.Sprintf("SELECT COUNT(*) FROM %s join %s on reading.device_info_id = device_info.id WHERE %s = false AND %s", readingTableName, deviceInfoTableName, markDeletedCol, whereCondition)
}

// sqlQueryCountByTimeRangeCol returns the SQL statement for counting the number of rows in the table
// by the given time range of the specified column
func sqlQueryCountByTimeRangeCol(table string, timeRangeCol string, arrayColNames []string, columns ...string) string {
	whereCondition := constructWhereNamedArgCondWithTimeRange(timeRangeCol, timeRangeCol, arrayColNames, columns...)
	return fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, whereCondition)
}

// sqlQueryCountEventByTimeRangeCol returns the SQL statement for counting the number of rows in the event table
// by the given time range of the specified column
func sqlQueryCountEventByTimeRangeCol(timeRangeCol string, arrayColNames []string, columns ...string) string {
	whereCondition := constructWhereNamedArgCondWithTimeRange(timeRangeCol, timeRangeCol, arrayColNames, columns...)
	return fmt.Sprintf("SELECT COUNT(*) FROM %s join %s on event.device_info_id = device_info.id WHERE %s = false AND %s", eventTableName, deviceInfoTableName, markDeletedCol, whereCondition)
}

// sqlQueryCountReadingByTimeRangeCol returns the SQL statement for counting the number of rows in the reading table
// by the given time range of the specified column
func sqlQueryCountReadingByTimeRangeCol(timeRangeCol string, arrayColNames []string, columns ...string) string {
	whereCondition := constructWhereNamedArgCondWithTimeRange(timeRangeCol, timeRangeCol, arrayColNames, columns...)
	return fmt.Sprintf("SELECT COUNT(*) FROM %s join %s on reading.device_info_id = device_info.id WHERE %s = false AND %s", readingTableName, deviceInfoTableName, markDeletedCol, whereCondition)
}

// sqlQueryCountByColAndLikePat returns the SQL statement for counting the number of rows by the given column name with LIKE pattern.
func sqlQueryCountByColAndLikePat(table string, columns ...string) string {
	whereCondition := constructWhereLikeCond(columns...)
	return fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, whereCondition)
}

// sqlQueryCountByJSONField returns the SQL statement for counting the number of rows in the table by the given JSON query string
func sqlQueryCountByJSONField(table string) string {
	return fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE content @> $1::jsonb", table)
}

// sqlQueryCountByTimeRange returns the SQL statement for counting the number of rows by the given time range
func sqlQueryCountByTimeRange(table string) string {
	return fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE COALESCE((content->>'%s')::bigint, 0) BETWEEN $1 AND $2", table, createdField)
}

func sqlQueryCountInUseResource() string {
	return fmt.Sprintf("SELECT count(resource) FROM %s device JOIN %s profile ON device.content->>'ProfileName'=profile.content->>'Name', jsonb_array_elements(profile.content->'DeviceResources') resource", deviceTableName, deviceProfileTableName)
}

func sqlCountEventByDeviceNameAndSourceNameAndLimit() string {
	return fmt.Sprintf(
		`SELECT count(*) FROM (
		  SELECT 1 FROM %s JOIN %s on event.device_info_id = device_info.id WHERE %s = false AND %s = @%s AND %s = @%s
		  LIMIT @%s
        ) limited_count`, eventTableName, deviceInfoTableName, markDeletedCol, deviceNameCol, deviceNameCol, sourceNameCol, sourceNameCol, limitCondition)
}

// sqlQueryEventIdFieldByTimeRangeAndConditions returns the SQL statement for selecting fields from the table within the time range
func sqlQueryEventIdFieldByTimeRangeAndConditions(timeRangeCol string, cols ...string) string {
	whereCondition := constructWhereNamedArgCondWithTimeRange("", timeRangeCol, nil, cols...)

	return fmt.Sprintf("SELECT event.id FROM %s JOIN %s on event.device_info_id = device_info.id WHERE %s", eventTableName, deviceInfoTableName, whereCondition)
}

// ----------------------------------------------------------------------------------
// SQL statements for UPDATE operations
// ----------------------------------------------------------------------------------

// sqlUpdateColsByCondCol returns the SQL statement for updating the passed columns of a row in the table by condCol.
func sqlUpdateColsByCondCol(table string, condCol string, cols ...string) string {
	columnCount := len(cols)
	updatedValues := constructUpdatedValues(cols...)
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s = $%d", table, updatedValues, condCol, columnCount+1)
}

// sqlUpdateColsByJSONCondCol returns the SQL statement for updating the passed columns of a row in the table by JSON field condition of content column.
func sqlUpdateColsByJSONCondCol(table string, cols ...string) string {
	columnCount := len(cols)
	updatedValues := constructUpdatedValues(cols...)
	return fmt.Sprintf("UPDATE %s SET %s WHERE content@>$%d::jsonb", table, updatedValues, columnCount+1)
}

// sqlUpdateContentById returns the SQL statement for updating the content of a row in the table by id.
func sqlUpdateContentById(table string) string {
	return fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2", table, contentCol, idCol)
}

// ----------------------------------------------------------------------------------
// SQL statements for DELETE operations
// ----------------------------------------------------------------------------------

// sqlDeleteById returns the SQL statement for deleting a row from the table by id.
func sqlDeleteById(table string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s = $1", table, idCol)
}

// sqlDeleteByAge returns the SQL statement for deleting rows from the table by created timestamp.
func sqlDeleteByAge(table string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s < NOW() - INTERVAL '1 millisecond' * $1", table, createdCol)
}

// sqlDeleteByContentAge returns the SQL statement for deleting rows from the table by created timestamp from content column.
func sqlDeleteByContentAge(table string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE COALESCE((content->>'%s')::bigint, 0) < (EXTRACT(EPOCH FROM NOW()) * 1000)::bigint - $1", table, createdField)
}

// sqlDeleteByContentAgeWithConds returns the SQL statement for deleting rows from the table by created timestamp from content column with the given where condition.
func sqlDeleteByContentAgeWithConds(table string, condition string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s AND COALESCE((content->>'%s')::bigint, 0) < (EXTRACT(EPOCH FROM NOW()) * 1000)::bigint - $1", table, condition, createdField)
}

// sqlDeleteByJSONField returns the SQL statement for deleting rows from the table by the given JSON query string
func sqlDeleteByJSONField(table string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE content @> $1::jsonb", table)
}

// sqlDeleteByJSONFieldAndAge returns the SQL statement for deleting rows from the table by the given column and created timestamp.
func sqlDeleteByJSONFieldAndAge(table string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE content @> $1::jsonb AND COALESCE((content->>'%s')::bigint, 0) < (EXTRACT(EPOCH FROM NOW()) * 1000)::bigint - $2", table, createdField)
}

// sqlDeleteTimeRangeByColumn returns the SQL statement for deleting rows from the table by time range with the specified column
// the time range is calculated from the caller function since the interval unit might be different
func sqlDeleteTimeRangeByColumn(table string, upperLimitTimeRangeCol string, cols ...string) string {
	whereCondition := constructWhereNamedArgCondWithTimeRange("", upperLimitTimeRangeCol, nil, cols...)
	return fmt.Sprintf("DELETE FROM %s WHERE %s", table, whereCondition)
}

// sqlDeleteEventsByTimeRangeAndColumn returns the SQL statement for deleting rows from the event table by time range with the specified column
// the time range is calculated from the caller function since the interval unit might be different
func sqlDeleteEventsByTimeRangeAndColumn(upperLimitTimeRangeCol string, cols ...string) string {
	whereCondition := constructWhereNamedArgCondWithTimeRange("", upperLimitTimeRangeCol, nil, cols...)
	return fmt.Sprintf("DELETE FROM %s USING %s WHERE event.device_info_id = device_info.id AND %s", eventTableName, deviceInfoTableName, whereCondition)
}

// sqlDeleteByColumn returns the SQL statement for deleting rows from the table by the specified column
func sqlDeleteByColumns(table string, cols ...string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s", table, constructWhereCondition(cols...))
}

// sqlDeleteEventsByColumn returns the SQL statement for deleting rows from the event table by the specified column
func sqlDeleteEventsByColumn(cols ...string) string {
	return fmt.Sprintf("DELETE FROM %s USING %s WHERE event.device_info_id = device_info.id AND %s", eventTableName, deviceInfoTableName, constructWhereNamedArgCondition(cols...))
}

// sqlDeleteByColAndLikePat returns the SQL statement for deleting rows by the specified column with LIKE pattern
// and append returnCol as result if not empty
func sqlDeleteByColAndLikePat(table string, column string, returnCol ...string) string {
	whereCond := constructWhereLikeCond(column)
	deleteStmt := fmt.Sprintf("DELETE FROM %s WHERE %s", table, whereCond)
	if len(returnCol) > 0 {
		deleteStmt += " RETURNING " + strings.Join(returnCol, ", ")
	}
	return deleteStmt
}

// ----------------------------------------------------------------------------------
// Utils
// ----------------------------------------------------------------------------------

// constructWhereCondition constructs the WHERE condition for the given columns.
func constructWhereCondition(columns ...string) string {
	columnCount := len(columns)
	conditions := make([]string, columnCount)

	for i, column := range columns {
		conditions[i] = fmt.Sprintf("%s = $%d", column, i+1)
	}

	return strings.Join(conditions, " AND ")
}

// constructWhereNamedArgCondition constructs the WHERE condition for the given columns.
func constructWhereNamedArgCondition(columns ...string) string {
	columnCount := len(columns)
	conditions := make([]string, columnCount)

	for i, column := range columns {
		conditions[i] = fmt.Sprintf("%s = @%s", column, column)
	}

	return strings.Join(conditions, " AND ")
}

// constructWhereNamedArgCondWithTimeRange constructs the WHERE condition for the given columns with time range
// if arrayColNames is not empty, ANY operator will be added to accept the array argument for the specified array col names
//
// For example, we want to query 'temperature' and 'humidity' from a 'sensor-device' within specified time range,
// we use origin as start, end time, and put deviceNameCol and resourceNameCol in columns and indicate the resourceNameCol is array in arrayColNames
// Here's the code snippet to query readings:
// sqlStatement := sqlQueryAllReadingWithPaginationAndTimeRangeDescByCol(originCol, originCol, []string{resourceNameCol}, deviceNameCol, resourceNameCol)
// queryArgs := pgx.NamedArgs{startTimeCondition: start, endTimeCondition: end, deviceNameCol: "sensor-device", resourceNameCol: []string{"temperature", "humidity"}, offsetCondition: 0, limitCondition: 100}
// readings, err := queryReadings(ctx, c.ConnPool, sqlStatement, queryArgs)
// ...
// The sqlQueryAllReadingWithPaginationAndTimeRangeDescByCol will invoke constructWhereNamedArgCondWithTimeRange to construct the where condition sql statement,
// the sql should similar to below:
// SELECT ... FROM ... WHERE origin >= @start AND origin <= @end AND devicename = @devicename AND resourcename = ANY (@resourcename) ORDER BY origin DESC OFFSET @offset LIMIT @limit
func constructWhereNamedArgCondWithTimeRange(startTimeCol, endTimeCol string, arrayColNames []string, columns ...string) string {
	var hasArrayColumn bool
	var conditions []string
	if startTimeCol != "" {
		conditions = append(conditions, fmt.Sprintf("%s >= @%s", startTimeCol, startTimeCondition))
	}
	if endTimeCol != "" {
		conditions = append(conditions, fmt.Sprintf("%s <= @%s", endTimeCol, endTimeCondition))
	}

	if len(arrayColNames) > 0 {
		hasArrayColumn = true
	}
	for _, column := range columns {
		equalCondition := "%s = @%s"
		if hasArrayColumn && slices.Contains(arrayColNames, column) {
			equalCondition = "%s = ANY (@%s)"
		}
		conditions = append(conditions, fmt.Sprintf(equalCondition, column, column))
	}

	return strings.Join(conditions, " AND ")
}

// constructWhereLikeCond constructs the WHERE condition for the given columns with LIKE operator
func constructWhereLikeCond(columns ...string) string {
	columnCount := len(columns)
	conditions := make([]string, columnCount)

	for i, column := range columns {
		conditions[i] = fmt.Sprintf("%s LIKE $%d", column, i+1)
	}

	return strings.Join(conditions, " AND ")
}

// constructWhereLikeCond constructs the updated values for SET keyword composed of the given columns
func constructUpdatedValues(columns ...string) string {
	columnCount := len(columns)
	conditions := make([]string, columnCount)

	for i, column := range columns {
		conditions[i] = fmt.Sprintf("%s = $%d", column, i+1)
	}

	return strings.Join(conditions, ", ")
}
