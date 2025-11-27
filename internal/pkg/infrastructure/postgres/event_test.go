//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/infrastructure/postgres/mocks"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteEvents(t *testing.T) {
	ctx := context.Background()
	sqlStatement := "DELETE FROM event WHERE id = @id"
	args := pgx.NamedArgs{"id": "test-id"}

	tests := []struct {
		name            string
		rowsAffected    int64
		execError       error
		expectError     bool
		expectedErrKind errors.ErrKind
		errorContains   []string
	}{
		{
			name:            "No rows affected - should return KindEntityDoesNotExist",
			rowsAffected:    0,
			execError:       nil,
			expectError:     true,
			expectedErrKind: errors.KindEntityDoesNotExist,
			errorContains:   []string{"no event found", "SQL statement:", sqlStatement},
		},
		{
			name:            "Rows affected - should succeed",
			rowsAffected:    1,
			execError:       nil,
			expectError:     false,
			expectedErrKind: "",
			errorContains:   nil,
		},
		{
			name:            "Exec returns error - should return KindDatabaseError",
			rowsAffected:    0,
			execError:       fmt.Errorf("database connection error"),
			expectError:     true,
			expectedErrKind: errors.KindDatabaseError,
			errorContains:   []string{"event(s) delete failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := new(mocks.Tx)
			// On successful completion, a DELETE command returns a command tag of the form "DELETE count"
			// https://www.postgresql.org/docs/16/sql-delete.html
			commandTag := pgconn.NewCommandTag(fmt.Sprintf("DELETE %d", tt.rowsAffected))
			tx.On("Exec", ctx, sqlStatement, args).Return(commandTag, tt.execError)

			err := deleteEvents(ctx, tx, sqlStatement, args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrKind, errors.Kind(err))
				for _, contains := range tt.errorContains {
					assert.Contains(t, err.Error(), contains)
				}
			} else {
				assert.NoError(t, err)
			}
			tx.AssertExpectations(t)
		})
	}
}

func TestDeleteReadingsBySubQuery(t *testing.T) {
	ctx := context.Background()
	subQuerySql := "SELECT id FROM event WHERE devicename = @devicename"
	args := pgx.NamedArgs{"devicename": "test-device"}

	tests := []struct {
		name            string
		rowsAffected    int64
		execError       error
		expectError     bool
		expectedErrKind errors.ErrKind
		errorContains   []string
	}{
		{
			name:            "No rows affected - should return KindEntityDoesNotExist",
			rowsAffected:    0,
			execError:       nil,
			expectError:     true,
			expectedErrKind: errors.KindEntityDoesNotExist,
			errorContains:   []string{"no reading found", "SQL statement:"},
		},
		{
			name:            "Rows affected - should succeed",
			rowsAffected:    1,
			execError:       nil,
			expectError:     false,
			expectedErrKind: "",
			errorContains:   nil,
		},
		{
			name:            "Exec returns error - should return KindDatabaseError",
			rowsAffected:    1, // Use non-zero to avoid RowsAffected check before error check
			execError:       fmt.Errorf("database connection error"),
			expectError:     true,
			expectedErrKind: errors.KindDatabaseError,
			errorContains:   []string{"reading(s) delete failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := new(mocks.Tx)
			// On successful completion, a DELETE command returns a command tag of the form "DELETE count"
			// https://www.postgresql.org/docs/16/sql-delete.html
			commandTag := pgconn.NewCommandTag(fmt.Sprintf("DELETE %d", tt.rowsAffected))
			tx.On("Exec", ctx, mock.MatchedBy(func(sql string) bool {
				return strings.Contains(sql, subQuerySql)
			}), args).Return(commandTag, tt.execError)

			err := deleteReadingsBySubQuery(ctx, tx, subQuerySql, args)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrKind, errors.Kind(err))
				for _, contains := range tt.errorContains {
					assert.Contains(t, err.Error(), contains)
				}
			} else {
				assert.NoError(t, err)
			}
			tx.AssertExpectations(t)
		})
	}
}
