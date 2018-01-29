//
// Copyright (c) 2017 Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import "go.uber.org/zap"

var logger *zap.Logger

// InitLogger - Init zap Logger
func InitLogger(l *zap.Logger) {
	logger = l
	return
}
