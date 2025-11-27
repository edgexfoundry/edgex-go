//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
)

// Tx is a mock implementation of pgx.Tx for testing
type Tx struct {
	mock.Mock
}

func (t *Tx) Begin(ctx context.Context) (pgx.Tx, error) {
	args := t.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (t *Tx) Commit(ctx context.Context) error {
	args := t.Called(ctx)
	return args.Error(0)
}

func (t *Tx) Rollback(ctx context.Context) error {
	args := t.Called(ctx)
	return args.Error(0)
}

func (t *Tx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	args := t.Called(ctx, tableName, columnNames, rowSrc)
	return args.Get(0).(int64), args.Error(1)
}

func (t *Tx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	args := t.Called(ctx, b)
	return args.Get(0).(pgx.BatchResults)
}

func (t *Tx) LargeObjects() pgx.LargeObjects {
	args := t.Called()
	return args.Get(0).(pgx.LargeObjects)
}

func (t *Tx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	args := t.Called(ctx, name, sql)
	return args.Get(0).(*pgconn.StatementDescription), args.Error(1)
}

func (t *Tx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	var callArgs []interface{}
	callArgs = []interface{}{ctx, sql}
	callArgs = append(callArgs, arguments...)
	mockArgs := t.Called(callArgs...)
	return mockArgs.Get(0).(pgconn.CommandTag), mockArgs.Error(1)
}

func (t *Tx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	mockArgs := t.Called(ctx, sql, args)
	return mockArgs.Get(0).(pgx.Rows), mockArgs.Error(1)
}

func (t *Tx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	mockArgs := t.Called(ctx, sql, args)
	return mockArgs.Get(0).(pgx.Row)
}

func (t *Tx) Conn() *pgx.Conn {
	args := t.Called()
	return args.Get(0).(*pgx.Conn)
}
