//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"

	"github.com/edgexfoundry/edgex-go/internal/core/keeper"

	"github.com/labstack/echo/v4"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	keeper.Main(ctx, cancel, echo.New(), os.Args[1:])
}
