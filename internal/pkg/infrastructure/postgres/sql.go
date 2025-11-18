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
	eventColumns   = "event.id, devicename, profilename, sourcename, origin, tags"
	readingColumns = "event_id, origin, value, binaryvalue, objectvalue, devicename, profilename, resourcename, valuetype, units, mediatype, tags"
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

// sqlQueryAll returns the SQL statement for selecting all rows from the table.
//func sqlQueryAll(table string) string {
//	return fmt.Sprintf("SELECT * FROM %s", table)
//}

// sqlQueryFieldsByCol returns the SQL statement for selecting the given fields of rows from the table by the conditions composed of given columns
func sqlQueryFieldsByCol(table string, fields []string, columns ...string) string {
	whereCondition := constructWhereCondition(columns...)
	queryFieldStr := strings.Join(fields, ", ")

	return fmt.Sprintf("SELECT %s FROM %s WHERE %s", queryFieldStr, table, whereCondition)
}

// sqlQueryEventIdFieldsByCol returns the SQL statement for selecting the event.id of rows from the event table by the conditions composed of given columns
func sqlQueryEventIdFieldsByCol(columns ...string) string {
	whereCondition := constructWhereCondition(columns...)

	return fmt.Sprintf("SELECT event.id FROM %s JOIN %s on event.device_info_id = device_info.id WHERE %s", eventTableName, deviceInfoTableName, whereCondition)
}

// sqlQueryFieldsByColAndLikePat returns the SQL statement for selecting the given fields of rows from the table by the conditions composed of given columns with LIKE operator
func sqlQueryFieldsByColAndLikePat(table string, fields []string, columns ...string) string {
	whereCondition := constructWhereLikeCond(columns...)
	queryFieldStr := strings.Join(fields, ", ")

	return fmt.Sprintf("SELECT %s FROM %s WHERE %s", queryFieldStr, table, whereCondition)
}

// sqlQueryAllWithTimeRange returns the SQL statement for selecting all rows from the table with a time range.
//func sqlQueryAllWithTimeRange(table string) string {
//	return fmt.Sprintf("SELECT * FROM %s WHERE %s >= $1 AND %s <= $2", table, createdCol, createdCol)
//}

// sqlQueryAllWithPaginationDesc returns the SQL statement for selecting all rows from the table with pagination by created timestamp in descending order.
//func sqlQueryAllWithPaginationDesc(table string) string {
//	return fmt.Sprintf("SELECT * FROM %s ORDER BY %s DESC OFFSET $1 LIMIT $2", table, createdCol)
//}

// sqlQueryAllWithConds returns the SQL statement for selecting all rows from the table by the given columns composed of the where condition
func sqlQueryAllWithConds(table string, columns ...string) string {
	whereCondition := constructWhereCondition(columns...)

	return fmt.Sprintf("SELECT * FROM %s WHERE %s", table, whereCondition)
}

// sqlQueryAllWithPaginationDescByCol returns the SQL statement for selecting all rows from the table with the pagination and desc by descCol
func sqlQueryAllWithPaginationDescByCol(table string, descCol string) string {
	return fmt.Sprintf("SELECT * FROM %s ORDER BY %s DESC OFFSET $1 LIMIT $2", table, descCol)
}

// sqlQueryAllEventWithPaginationDescByCol returns the SQL statement for selecting all rows from the event table with the pagination and desc by descCol
func sqlQueryAllEventWithPaginationDescByCol(descCol string) string {
	return fmt.Sprintf("SELECT %s FROM %s JOIN %s on event.device_info_id = device_info.id WHERE %s = false ORDER BY %s DESC OFFSET $1 LIMIT $2", eventColumns, eventTableName, deviceInfoTableName, markDeletedCol, descCol)
}

// sqlQueryAllReadingWithPaginationDescByCol returns the SQL statement for selecting all rows from the reading table with the pagination and desc by descCol
func sqlQueryAllReadingWithPaginationDescByCol(descCol string) string {
	return fmt.Sprintf("SELECT %s FROM %s JOIN %s on reading.device_info_id = device_info.id WHERE %s = false ORDER BY %s DESC OFFSET $1 LIMIT $2", readingColumns, readingTableName, deviceInfoTableName, markDeletedCol, descCol)
}

// sqlQueryAllReadingAndDescWithConds returns the SQL statement for selecting all rows from the table by the given columns composed of the where condition with descending by descCol
func sqlQueryAllReadingAndDescWithConds(descCol string, columns ...string) string {
	whereCondition := constructWhereCondition(columns...)

	return fmt.Sprintf("SELECT %s FROM %s JOIN %s on reading.device_info_id = device_info.id WHERE %s = false AND %s ORDER BY %s DESC", readingColumns, readingTableName, deviceInfoTableName, markDeletedCol, whereCondition, descCol)
}

// sqlQueryAllEventAndDescWithCondsAndPage returns the SQL statement for selecting all rows from the event table by the given columns composed of the where condition
// with descending by descCol and pagination
func sqlQueryAllEventAndDescWithCondsAndPage(descCol string, columns ...string) string {
	columnCount := len(columns)
	whereCondition := constructWhereCondition(columns...)

	return fmt.Sprintf(
		"SELECT %s FROM %s join %s on event.device_info_id = device_info.id WHERE %s = false AND %s ORDER BY %s DESC OFFSET $%d LIMIT $%d",
		eventColumns, eventTableName, deviceInfoTableName, markDeletedCol,
		whereCondition, descCol,
		// note that this is a prepared statement with parameters beginning with count WHERE
		// conditions, so adding 1 and 2 for OFFSET, LIMIT parameters, respectively
		columnCount+1, columnCount+2)
}

// sqlQueryAllReadingAndDescWithCondsAndPag returns the SQL statement for selecting all rows from the reading table by the given columns composed of the where condition
// with descending by descCol and pagination
func sqlQueryAllReadingAndDescWithCondsAndPag(descCol string, columns ...string) string {
	columnCount := len(columns)
	whereCondition := constructWhereCondition(columns...)

	return fmt.Sprintf(
		"SELECT %s FROM %s join %s on reading.device_info_id = device_info.id WHERE %s = false AND %s ORDER BY %s DESC OFFSET $%d LIMIT $%d",
		readingColumns, readingTableName, deviceInfoTableName, markDeletedCol,
		whereCondition, descCol,
		// note that this is a prepared statement with parameters beginning with count WHERE
		// conditions, so adding 1 and 2 for OFFSET, LIMIT parameters, respectively
		columnCount+1, columnCount+2)
}

// sqlQueryAllEventAndDescWithCondsAndPagAndUpperLimitTime returns the SQL statement for selecting all rows from the event table by the given columns composed of the where condition
// with descending by descCol and pagination
func sqlQueryAllEventAndDescWithCondsAndPagAndUpperLimitTime(descCol string, upperLimitTimeRangeCol string, columns ...string) string {
	columnCount := len(columns)
	whereCondition := constructWhereCondWithTimeRange("", upperLimitTimeRangeCol, nil, columns...)

	return fmt.Sprintf(
		"SELECT %s FROM %s join %s on event.device_info_id = device_info.id WHERE %s = false AND %s ORDER BY %s DESC OFFSET $%d LIMIT $%d",
		eventColumns, eventTableName, deviceInfoTableName, markDeletedCol,
		whereCondition, descCol,
		// note that this is a prepared statement with parameters beginning with UpperLimitTime
		// and then columns conditions, so adding 2 and 3 for OFFSET, LIMIT parameters, respectively
		columnCount+2, columnCount+3)
}

// sqlQueryAllWithPaginationAndTimeRange returns the SQL statement for selecting all rows from the table with pagination and a time range.
func sqlQueryAllWithPaginationAndTimeRange(table string) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE %s >= $1 AND %s <= $2 ORDER BY %s OFFSET $3 LIMIT $4", table, createdCol, createdCol, createdCol)
}

// sqlQueryAllEventWithPaginationAndTimeRangeDescByCol returns the SQL statement for selecting all rows from the event table with the arrayColNames slice,
// provided columns with pagination and a time range by timeRangeCol, desc by descCol
func sqlQueryAllEventWithPaginationAndTimeRangeDescByCol(timeRangeCol string, descCol string, arrayColNames []string, columns ...string) string {
	whereCondition := constructWhereCondWithTimeRange(timeRangeCol, timeRangeCol, arrayColNames, columns...)
	columnCount := len(columns)

	return fmt.Sprintf(
		"SELECT %s FROM %s join %s on event.device_info_id = device_info.id WHERE %s = false AND %s ORDER BY %s DESC OFFSET $%d LIMIT $%d",
		eventColumns, eventTableName, deviceInfoTableName, markDeletedCol,
		whereCondition, descCol,
		// note that this is a prepared statement with parameters beginning with two timeRangeCol
		// and then columns conditions, so OFFSET, LIMIT are the third and forth parameters
		columnCount+3, columnCount+4)
}

// sqlQueryAllReadingWithPaginationAndTimeRangeDescByCol returns the SQL statement for selecting all rows from the reading table with the arrayColNames slice,
// provided columns with pagination and a time range by timeRangeCol, desc by descCol
func sqlQueryAllReadingWithPaginationAndTimeRangeDescByCol(timeRangeCol string, descCol string, arrayColNames []string, columns ...string) string {
	whereCondition := constructWhereCondWithTimeRange(timeRangeCol, timeRangeCol, arrayColNames, columns...)
	columnCount := len(columns)

	return fmt.Sprintf(
		"SELECT %s FROM %s join %s on reading.device_info_id = device_info.id WHERE %s = false AND %s ORDER BY %s DESC OFFSET $%d LIMIT $%d",
		readingColumns, readingTableName, deviceInfoTableName, markDeletedCol,
		whereCondition, descCol,
		// note that this is a prepared statement with parameters beginning with two timeRangeCol
		// and then columns conditions, so OFFSET, LIMIT are the third and forth parameters
		columnCount+3, columnCount+4)
}

// sqlQueryAllByStatusWithPaginationAndTimeRange returns the SQL statement for selecting all rows from the table by status with pagination and a time range.
func sqlQueryAllByStatusWithPaginationAndTimeRange(table string) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE %s = $1 AND %s >= $2 AND %s <= $3 ORDER BY %s OFFSET $4 LIMIT $5", table, statusCol, createdCol, createdCol, createdCol)
}

// sqlQueryAllByColWithPaginationAndTimeRange returns the SQL statement for selecting all rows from the table by the given columns with pagination and a time range.
func sqlQueryAllByColWithPaginationAndTimeRange(table string, columns ...string) string {
	columnCount := len(columns)
	whereCondition := constructWhereCondition(columns...)
	timeRangeCondition := fmt.Sprintf("%s >= $%d AND %s <= $%d", createdCol, columnCount+1, createdCol, columnCount+2)

	return fmt.Sprintf(
		"SELECT * FROM %s WHERE %s AND %s ORDER BY %s OFFSET $%d LIMIT $%d",
		table, whereCondition, timeRangeCondition, createdCol,
		// note that this is a prepared statement with parameters beginning with two timeRangeCol
		// and then columns conditions, so OFFSET, LIMIT are the third and forth parameters
		columnCount+3, columnCount+4)
}

// sqlQueryAllWithPaginationAndTimeRangeDesc returns the SQL statement for selecting all rows from the table with pagination and a time range.
//func sqlQueryAllWithPaginationAndTimeRangeDesc(table string) string {
//	return fmt.Sprintf("SELECT * FROM %s WHERE %s >= $1 AND %s <= $2 ORDER BY %s DESC OFFSET $3 LIMIT $4", table, createdCol, createdCol, createdCol)
//}

// sqlQueryAllById returns the SQL statement for selecting all rows from the table by id.
func sqlQueryAllById(table string) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE %s = $1", table, idCol)
}

// sqlQueryAllEventById returns the SQL statement for selecting all rows from the event table by id.
func sqlQueryAllEventById() string {
	return fmt.Sprintf("SELECT %s FROM %s JOIN %s on event.device_info_id = device_info.id WHERE core_data.event.id=$1", eventColumns, eventTableName, deviceInfoTableName)
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

// sqlQueryContentWithTimeRangeAndPagination returns the SQL statement for selecting content column from the table by the given time range with pagination
func sqlQueryContentWithTimeRangeAndPagination(table string) string {
	return fmt.Sprintf("SELECT content FROM %s WHERE COALESCE((content->>'%s')::bigint, 0) BETWEEN $1 AND $2 AND content @> $3::jsonb ORDER BY COALESCE((content->>'%s')::bigint, 0) OFFSET $4 LIMIT $5", table, createdField, createdField)
}

// sqlQueryContentByJSONField returns the SQL statement for selecting content column in the table by the given JSON query string
func sqlQueryContentByJSONField(table string) string {
	return fmt.Sprintf("SELECT content FROM %s WHERE content @> $1::jsonb", table)
}

// sqlQueryContentByJSONFieldWithPagination returns the SQL statement for selecting content column in the table by the given JSON query string with pagination
func sqlQueryContentByJSONFieldWithPagination(table string) string {
	return fmt.Sprintf("SELECT content FROM %s WHERE content @> $1::jsonb ORDER BY COALESCE((content->>'%s')::bigint, 0) OFFSET $2 LIMIT $3", table, createdField)
}

// sqlQueryContentByJSONFieldTimeRange returns the SQL statement for selecting content column by the given time range of the JSON field name
//func sqlQueryContentByJSONFieldTimeRange(table string, field string) string {
//	return fmt.Sprintf("SELECT content FROM %s WHERE (content->'%s')::bigint  >= $1 AND (content->'%s')::bigint <= $2 ORDER BY %s OFFSET $3 LIMIT $4", table, field, field, createdCol)
//}

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
	whereCondition := constructWhereCondition(columns...)
	return fmt.Sprintf("SELECT COUNT(*) FROM %s join %s on event.device_info_id = device_info.id WHERE %s = false AND %s", eventTableName, deviceInfoTableName, markDeletedCol, whereCondition)
}

// sqlQueryCountReadingByCol returns the SQL statement for counting the number of rows in the table by the given column name.
func sqlQueryCountReadingByCol(columns ...string) string {
	whereCondition := constructWhereCondition(columns...)
	return fmt.Sprintf("SELECT COUNT(*) FROM %s join %s on reading.device_info_id = device_info.id WHERE %s = false AND %s", readingTableName, deviceInfoTableName, markDeletedCol, whereCondition)
}

// sqlQueryCountByTimeRangeCol returns the SQL statement for counting the number of rows in the table
// by the given time range of the specified column
func sqlQueryCountByTimeRangeCol(table string, timeRangeCol string, arrayColNames []string, columns ...string) string {
	whereCondition := constructWhereCondWithTimeRange(timeRangeCol, timeRangeCol, arrayColNames, columns...)
	return fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, whereCondition)
}

// sqlQueryCountEventByTimeRangeCol returns the SQL statement for counting the number of rows in the event table
// by the given time range of the specified column
func sqlQueryCountEventByTimeRangeCol(timeRangeCol string, arrayColNames []string, columns ...string) string {
	whereCondition := constructWhereCondWithTimeRange(timeRangeCol, timeRangeCol, arrayColNames, columns...)
	return fmt.Sprintf("SELECT COUNT(*) FROM %s join %s on event.device_info_id = device_info.id WHERE %s = false AND %s", eventTableName, deviceInfoTableName, markDeletedCol, whereCondition)
}

// sqlQueryCountReadingByTimeRangeCol returns the SQL statement for counting the number of rows in the reading table
// by the given time range of the specified column
func sqlQueryCountReadingByTimeRangeCol(timeRangeCol string, arrayColNames []string, columns ...string) string {
	whereCondition := constructWhereCondWithTimeRange(timeRangeCol, timeRangeCol, arrayColNames, columns...)
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
		  SELECT 1 FROM %s JOIN %s on event.device_info_id = device_info.id WHERE %s = false AND %s = $1 AND %s = $2
		  LIMIT $3
        ) limited_count`, eventTableName, deviceInfoTableName, markDeletedCol, deviceNameCol, sourceNameCol)
}

// sqlQueryCountByJSONFieldTimeRange returns the SQL statement for counting the number of rows in the table
// by the given time range of the JSON field name
//func sqlQueryCountByJSONFieldTimeRange(table string, field string) string {
//	return fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE (content->'%s')::bigint  >= $1 AND (content->'%s')::bigint <= $2", table, field, field)
//}

// sqlQueryEventIdFieldByTimeRangeAndConditions returns the SQL statement for selecting fields from the table within the time range
func sqlQueryEventIdFieldByTimeRangeAndConditions(timeRangeCol string, cols ...string) string {
	whereCondition := constructWhereCondWithTimeRange("", timeRangeCol, nil, cols...)

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
	whereCondition := constructWhereCondWithTimeRange("", upperLimitTimeRangeCol, nil, cols...)
	return fmt.Sprintf("DELETE FROM %s WHERE %s", table, whereCondition)
}

// sqlDeleteEventsByTimeRangeAndColumn returns the SQL statement for deleting rows from the event table by time range with the specified column
// the time range is calculated from the caller function since the interval unit might be different
func sqlDeleteEventsByTimeRangeAndColumn(upperLimitTimeRangeCol string, cols ...string) string {
	whereCondition := constructWhereCondWithTimeRange("", upperLimitTimeRangeCol, nil, cols...)
	return fmt.Sprintf("DELETE FROM %s USING %s WHERE event.device_info_id = device_info.id AND %s", eventTableName, deviceInfoTableName, whereCondition)
}

// sqlDeleteByColumn returns the SQL statement for deleting rows from the table by the specified column
func sqlDeleteByColumns(table string, cols ...string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s", table, constructWhereCondition(cols...))
}

// sqlDeleteEventsByColumn returns the SQL statement for deleting rows from the event table by the specified column
func sqlDeleteEventsByColumn(cols ...string) string {
	return fmt.Sprintf("DELETE FROM %s USING %s WHERE event.device_info_id = device_info.id AND %s", eventTableName, deviceInfoTableName, constructWhereCondition(cols...))
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

// constructWhereCondWithTimeRange constructs the WHERE condition for the given columns with time range
// if arrayColNames is not empty, ANY operator will be added to accept the array argument for the specified array col names
func constructWhereCondWithTimeRange(lowerLimitTimeRangeCol, upperLimitTimeRangeCol string, arrayColNames []string, columns ...string) string {
	var hasArrayColumn bool
	var conditions []string
	if lowerLimitTimeRangeCol != "" {
		conditions = append(conditions, lowerLimitTimeRangeCol+" >= $1")
	}
	if upperLimitTimeRangeCol != "" {
		conditions = append(conditions, fmt.Sprintf("%s <= $%d", upperLimitTimeRangeCol, len(conditions)+1))
	}

	if len(arrayColNames) > 0 {
		hasArrayColumn = true
	}
	for _, column := range columns {
		equalCondition := "%s = $%d"
		if hasArrayColumn && slices.Contains(arrayColNames, column) {
			equalCondition = "%s = ANY ($%d)"
		}
		conditions = append(conditions, fmt.Sprintf(equalCondition, column, 1+len(conditions)))
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
