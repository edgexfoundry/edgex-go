//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

type AdvisoryLock interface {
	Lock() (bool, error)
	Unlock() (bool, error)
	LockShared() (bool, error)
	UnlockShared() (bool, error)
}

// NewAdvisoryLock creates a new instance with specified lock ID, and connection pool.
func NewAdvisoryLock(connPool *pgxpool.Pool, logger logger.LoggingClient, serviceKey string) (AdvisoryLock, error) {
	lockID := generateHashLockId(serviceKey)
	logger.Debugf("Use Advisory lock ID: %d for service %s", lockID, serviceKey)
	return &pgAdvisoryLock{
		logger:   logger,
		connPool: connPool,
		lockID:   lockID,
	}, nil
}

// pgAdvisoryLock is a struct realization of PostgreSQL AdvisoryLock that holds the connection pool and lock ID for the
// advisory lock.
type pgAdvisoryLock struct {
	logger   logger.LoggingClient
	connPool *pgxpool.Pool
	lockID   int64
	mutex    sync.RWMutex
}

func (l *pgAdvisoryLock) callLockFunction(query string) (bool, error) {
	if l.connPool == nil {
		return false, fmt.Errorf("connection pool is nil")
	}
	return lock(context.Background(), l.connPool, query, l.lockID)
}

// Lock tries to acquire an exclusive lock on the lock ID. If the lock is acquired, it returns true; otherwise, it
// returns false.
func (l *pgAdvisoryLock) Lock() (bool, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.callLockFunction("SELECT pg_try_advisory_lock($1)")
}

// Unlock releases the exclusive lock on the lock ID. If the lock is released, it returns true; otherwise, it returns
// false.
func (l *pgAdvisoryLock) Unlock() (bool, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.callLockFunction("SELECT pg_advisory_unlock($1)")
}

// LockShared tries to acquire a shared lock on the lock ID. If the lock is acquired, it returns true; otherwise, it
// returns false.
func (l *pgAdvisoryLock) LockShared() (bool, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.callLockFunction("SELECT pg_try_advisory_lock_shared($1)")
}

// UnlockShared releases the shared lock on the lock ID. If the lock is released, it returns true; otherwise, it returns
// false.
func (l *pgAdvisoryLock) UnlockShared() (bool, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.callLockFunction("SELECT pg_advisory_unlock_shared($1)")
}

func lock(ctx context.Context, connPool *pgxpool.Pool, query string, lockId int64) (result bool, err error) {
	if connPool == nil {
		return false, fmt.Errorf("connection pool is nil")
	}
	err = connPool.QueryRow(ctx, query, lockId).Scan(&result)
	if err != nil {
		return false, fmt.Errorf("error while trying to read response rows from locking function: %w", err)
	}
	return result, nil
}

// generateHashLockId is a handy func to generates a hash value from the service key (string) and returns the int64 value.
// note that the hash value is generated using FNV-1a algorithm, which is a non-cryptographic hash function intended for
// applications that do not need the rigorous security requirement and can be faster and less resources-intensive.
func generateHashLockId(serviceKey string) int64 {
	// generate a hash value from the string using Hash32a algorithm as the pg_try_advisory_lock function takes a
	// bigint with value range from -2^63 to 2^63-1, which exactly matches the int64 value range. However, the 64-bit
	// FNV-1a hash can only reproduce uint64 value that exceeds the bigint value range, so we use the 32-bit FNV-1a hash
	// instead.
	h := fnv.New32a()
	h.Write([]byte(serviceKey))
	return int64(h.Sum32())
}
