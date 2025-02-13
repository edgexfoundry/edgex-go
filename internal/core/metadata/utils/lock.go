//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import "sync"

type CapacityCheckLock struct {
	mutex sync.RWMutex
}

func NewCapacityCheckLock() *CapacityCheckLock {
	return &CapacityCheckLock{}
}

func (c *CapacityCheckLock) Lock() {
	c.mutex.Lock()
}
func (c *CapacityCheckLock) Unlock() {
	c.mutex.Unlock()
}
