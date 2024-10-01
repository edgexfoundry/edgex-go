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

// sqlQueryAllWithPaginationDescByCol returns the SQL statement for selecting all rows from the table with the pagination and desc by descCol
func sqlQueryAllWithPaginationDescByCol(table string, descCol string) string {
	return fmt.Sprintf("SELECT * FROM %s ORDER BY %s DESC OFFSET $1 LIMIT $2", table, descCol)
}

// sqlQueryAllAndDescWithConds returns the SQL statement for selecting all rows from the table by the given columns composed of the where condition with descending by descCol
func sqlQueryAllAndDescWithConds(table string, descCol string, columns ...string) string {
	whereCondition := constructWhereCondition(columns...)

	return fmt.Sprintf("SELECT * FROM %s WHERE %s ORDER BY %s DESC", table, whereCondition, descCol)
}

// sqlQueryAllAndDescWithCondsAndPag returns the SQL statement for selecting all rows from the table by the given columns composed of the where condition
// with descending by descCol and pagination
func sqlQueryAllAndDescWithCondsAndPag(table string, descCol string, columns ...string) string {
	columnCount := len(columns)
	whereCondition := constructWhereCondition(columns...)

	return fmt.Sprintf("SELECT * FROM %s WHERE %s ORDER BY %s DESC OFFSET $%d LIMIT $%d", table, whereCondition, descCol, columnCount+1, columnCount+2)
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

// sqlQueryAllById returns the SQL statement for selecting all rows from the table by id.
func sqlQueryAllById(table string) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE %s = $1", table, idCol)
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
	return fmt.Sprintf("SELECT content FROM %s WHERE COALESCE((content->>'%s')::bigint, 0) BETWEEN $1 AND $2 ORDER BY COALESCE((content->>'%s')::bigint, 0) OFFSET $3 LIMIT $4", table, createdField, createdField)
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
func sqlDeleteTimeRangeByColumn(table string, column string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s <= $1", table, column)
}

// sqlDeleteByColumn returns the SQL statement for deleting rows from the table by the specified column
func sqlDeleteByColumn(table string, column string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s = $1", table, column)
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