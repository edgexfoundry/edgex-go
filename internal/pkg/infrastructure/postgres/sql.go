//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This file contains the common SQL statements for interacting with the PostgreSQL database.
// All the arguments need to manually pass to the actual jackc/pgx operation functions.

package postgres

import (
	"fmt"
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

// sqlQueryAllWithTimeRange returns the SQL statement for selecting all rows from the table with a time range.
//func sqlQueryAllWithTimeRange(table string) string {
//	return fmt.Sprintf("SELECT * FROM %s WHERE %s >= $1 AND %s <= $2", table, createdCol, createdCol)
//}

// sqlQueryAllWithPagination returns the SQL statement for selecting all rows from the table with pagination.
//func sqlQueryAllWithPagination(table string) string {
//	return fmt.Sprintf("SELECT * FROM %s OFFSET $1 LIMIT $2", table)
//}

// sqlQueryAllWithPaginationDesc returns the SQL statement for selecting all rows from the table with pagination by created timestamp in descending order.
//func sqlQueryAllWithPaginationDesc(table string) string {
//	return fmt.Sprintf("SELECT * FROM %s ORDER BY %s DESC OFFSET $1 LIMIT $2", table, createdCol)
//}

// sqlQueryAllWithPaginationAndTimeRange returns the SQL statement for selecting all rows from the table with pagination and a time range.
func sqlQueryAllWithPaginationAndTimeRange(table string) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE %s >= $1 AND %s <= $2 ORDER BY %s OFFSET $3 LIMIT $4", table, createdCol, createdCol, createdCol)
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
	return fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = $1", table, whereCondition)
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
//func sqlDeleteById(table string) string {
//	return fmt.Sprintf("DELETE FROM %s WHERE %s = $1", table, idCol)
//}

// sqlDeleteByAge returns the SQL statement for deleting rows from the table by created timestamp.
func sqlDeleteByAge(table string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s < NOW() - INTERVAL '1 millisecond' * $1", table, createdCol)
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
