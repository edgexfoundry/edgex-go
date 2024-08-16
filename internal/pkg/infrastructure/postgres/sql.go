//
// Copyright (C) 2024 IOTech Ltd
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

// Constants for common column names in the database.
const (
	contentCol  = "content"
	createdCol  = "created"
	idCol       = "id"
	modifiedCol = "modified"
	nameCol     = "name"
	statusCol   = "status"
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

// sqlInsertContent returns the SQL statement for inserting a new row with id, name, and content into the table.
func sqlInsertContent(table string) string {
	return fmt.Sprintf("INSERT INTO %s(%s, %s, %s) VALUES ($1, $2, $3)", table, idCol, nameCol, contentCol)
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

// sqlQueryAllWithTimeRange returns the SQL statement for selecting all rows from the table with a time range.
//func sqlQueryAllWithTimeRange(table string) string {
//	return fmt.Sprintf("SELECT * FROM %s WHERE %s >= $1 AND %s <= $2", table, createdCol, createdCol)
//}

// sqlQueryAllWithPagination returns the SQL statement for selecting all rows from the table with pagination.
func sqlQueryAllWithPagination(table string) string {
	return fmt.Sprintf("SELECT * FROM %s OFFSET $1 LIMIT $2", table)
}

// sqlQueryAllWithPaginationDesc returns the SQL statement for selecting all rows from the table with pagination by created timestamp in descending order.
//func sqlQueryAllWithPaginationDesc(table string) string {
//	return fmt.Sprintf("SELECT * FROM %s ORDER BY %s DESC OFFSET $1 LIMIT $2", table, createdCol)
//}

// sqlQueryAllWithPaginationDescByCol returns the SQL statement for selecting all rows from the table with the pagination and desc by descCol
func sqlQueryAllWithPaginationDescByCol(table string, descCol string) string {
	return fmt.Sprintf("SELECT * FROM %s ORDER BY %s DESC OFFSET $1 LIMIT $2", table, descCol)
}

// sqlQueryAllByColWithPagination returns the SQL statement for selecting all rows from the table by the given columns with pagination
func sqlQueryAllByColWithPagination(table string, columns ...string) string {
	columnCount := len(columns)
	whereCondition := constructWhereCondition(columns...)

	return fmt.Sprintf("SELECT * FROM %s WHERE %s OFFSET $%d LIMIT $%d", table, whereCondition, columnCount+1, columnCount+2)
}

// sqlQueryAllWithPaginationAndTimeRange returns the SQL statement for selecting all rows from the table with pagination and a time range.
func sqlQueryAllWithPaginationAndTimeRange(table string) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE %s >= $1 AND %s <= $2 ORDER BY %s OFFSET $3 LIMIT $4", table, createdCol, createdCol, createdCol)
}

// sqlQueryAllWithPaginationAndTimeRangeDescByCol returns the SQL statement for selecting all rows from the table with the arrayColNames slice,
// provided columns with pagination and a time range by timeRangeCol, desc by descCol
func sqlQueryAllWithPaginationAndTimeRangeDescByCol(table string, timeRangeCol string, descCol string, arrayColNames []string, columns ...string) string {
	whereCondition := constructWhereCondWithTimeRange(timeRangeCol, arrayColNames, columns...)
	columnCount := len(columns)

	return fmt.Sprintf("SELECT * FROM %s WHERE %s ORDER BY %s DESC OFFSET $%d LIMIT $%d",
		table, whereCondition, descCol, columnCount+3, columnCount+4)
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

	return fmt.Sprintf("SELECT * FROM %s WHERE %s AND %s ORDER BY %s OFFSET $%d LIMIT $%d", table, whereCondition, timeRangeCondition, createdCol, columnCount+3, columnCount+4)
}

// sqlQueryAllWithPaginationAndTimeRangeDesc returns the SQL statement for selecting all rows from the table with pagination and a time range.
//func sqlQueryAllWithPaginationAndTimeRangeDesc(table string) string {
//	return fmt.Sprintf("SELECT * FROM %s WHERE %s >= $1 AND %s <= $2 ORDER BY %s DESC OFFSET $3 LIMIT $4", table, createdCol, createdCol, createdCol)
//}

// sqlQueryAllByName returns the SQL statement for selecting all rows from the table by name.
func sqlQueryAllByName(table string) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE %s = $1", table, nameCol)
}

// sqlQueryAllById returns the SQL statement for selecting all rows from the table by id.
func sqlQueryAllById(table string) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE %s = $1", table, idCol)
}

// sqlQueryAllById returns the SQL statement for selecting content column by the specified id.
//func sqlQueryContentById(table string) string {
//	return fmt.Sprintf("SELECT content FROM %s WHERE %s = $1", table, idCol)
//}

// sqlQueryContentWithPagination returns the SQL statement for selecting content column from the table with pagination
//func sqlQueryContentWithPagination(table string) string {
//	return fmt.Sprintf("SELECT content FROM %s ORDER BY created OFFSET $1 LIMIT $2", table)
//}

// sqlQueryCountByJSONField returns the SQL statement for selecting content column in the table by the given JSON query string
//func sqlQueryContentByJSONField(table string) (string, errors.EdgeX) {
//	return fmt.Sprintf("SELECT content FROM %s WHERE content @> $1::jsonb ORDER BY %s OFFSET $2 LIMIT $3", table, createdCol), nil
//}

// sqlQueryCountByJSONField returns the SQL statement for selecting content column by the given time range of the JSON field name
//func sqlQueryContentByJSONFieldTimeRange(table string, field string) string {
//	return fmt.Sprintf("SELECT content FROM %s WHERE (content->'%s')::bigint  >= $1 AND (content->'%s')::bigint <= $2 ORDER BY %s OFFSET $3 LIMIT $4", table, field, field, createdCol)
//}

// sqlCheckExistsByName returns the SQL statement for checking if a row exists in the table by name.
func sqlCheckExistsByName(table string) string {
	return fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = $1)", table, nameCol)
}

// sqlCheckExistsById returns the SQL statement for checking if a row exists in the table by id.
//func sqlCheckExistsById(table string) string {
//	return fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = $1)", table, idCol)
//}

// sqlQueryCount returns the SQL statement for counting the number of rows in the table.
func sqlQueryCount(table string) string {
	return fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
}

// sqlQueryCountByCol returns the SQL statement for counting the number of rows in the table by the given column name.
func sqlQueryCountByCol(table string, columns ...string) string {
	whereCondition := constructWhereCondition(columns...)
	return fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, whereCondition)
}

// sqlQueryCountByTimeRangeCol returns the SQL statement for counting the number of rows in the table
// by the given time range of the specified column
func sqlQueryCountByTimeRangeCol(table string, timeRangeCol string, arrayColNames []string, columns ...string) string {
	whereCondition := constructWhereCondWithTimeRange(timeRangeCol, arrayColNames, columns...)
	return fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, whereCondition)
}

// sqlQueryCountByJSONField returns the SQL statement for counting the number of rows in the table by the given JSON query string
//func sqlQueryCountByJSONField(table string) (string, errors.EdgeX) {
//	return fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE content @> $1::jsonb", table), nil
//}

// sqlQueryCountByJSONFieldTimeRange returns the SQL statement for counting the number of rows in the table
// by the given time range of the JSON field name
//func sqlQueryCountByJSONFieldTimeRange(table string, field string) string {
//	return fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE (content->'%s')::bigint  >= $1 AND (content->'%s')::bigint <= $2", table, field, field)
//}

// sqlQueryFieldsByTimeRange returns the SQL statement for selecting fields from the table within the time range
func sqlQueryFieldsByTimeRange(table string, fields []string, timeRangeCol string) string {
	queryFieldStr := strings.Join(fields, ", ")

	return fmt.Sprintf("SELECT %s FROM %s WHERE %s <= $1", queryFieldStr, table, timeRangeCol)
}

// ----------------------------------------------------------------------------------
// SQL statements for UPDATE operations
// ----------------------------------------------------------------------------------

// sqlUpdateContentByName returns the SQL statement for updating the content and modified timestamp of a row in the table by name.
func sqlUpdateContentByName(table string) string {
	return fmt.Sprintf("UPDATE %s SET %s = $1 , %s = $2 WHERE %s = $3", table, contentCol, modifiedCol, nameCol)
}

// sqlUpdateContentById returns the SQL statement for updating the content and modified timestamp of a row in the table by id.
//func sqlUpdateContentById(table string) string {
//	return fmt.Sprintf("UPDATE %s SET %s = $1 , %s = $2 WHERE %s = $3", table, contentCol, modifiedCol, idCol)
//}

// ----------------------------------------------------------------------------------
// SQL statements for DELETE operations
// ----------------------------------------------------------------------------------

// sqlDeleteByName returns the SQL statement for deleting a row from the table by name.
func sqlDeleteByName(table string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s = $1", table, nameCol)
}

// sqlDeleteById returns the SQL statement for deleting a row from the table by id.
func sqlDeleteById(table string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s = $1", table, idCol)
}

// sqlDeleteByAge returns the SQL statement for deleting rows from the table by created timestamp.
func sqlDeleteByAge(table string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s < NOW() - INTERVAL '1 millisecond' * $1", table, createdCol)
}

// sqlDeleteTimeRangeByColumn returns the SQL statement for deleting rows from the table by time range with the specified column
// the time range is calculated from the caller function since the interval unit might be different
func sqlDeleteTimeRangeByColumn(table string, column string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s <= $1", table, column)
}

// sqlDeleteByColumn returns the SQL statement for deleting rows from the table by the specified column
func sqlDeleteByColumn(table string, column string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s = $1", table, column)
}

// sqlDeleteByJSONField returns the SQL statement for deleting rows from the table by the given JSON query string
//func sqlDeleteByJSONField(table string) (string, errors.EdgeX) {
//	return fmt.Sprintf("DELETE FROM %s WHERE content @> $1::jsonb", table), nil
//}

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

// constructWhereCondition constructs the WHERE condition for the given columns with time range
// if arrayColNames is not empty, ANY operator will be added to accept the array argument for the specified array col names
func constructWhereCondWithTimeRange(timeRangeCol string, arrayColNames []string, columns ...string) string {
	var hasArrayColumn bool
	conditions := []string{timeRangeCol + " >= $1", timeRangeCol + " <= $2"}

	if len(arrayColNames) > 0 {
		hasArrayColumn = true
	}
	for i, column := range columns {
		equalCondition := "%s = $%d"
		if hasArrayColumn && slices.Contains(arrayColNames, column) {
			equalCondition = "%s = ANY ($%d)"
		}
		conditions = append(conditions, fmt.Sprintf(equalCondition, column, i+3))
	}

	return strings.Join(conditions, " AND ")
}
