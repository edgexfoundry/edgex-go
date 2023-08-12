/*******************************************************************************
 * Copyright 2018 Dell Inc.
 * Copyright (C) 2023 IOTech Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package main

import (
	"context"

	"github.com/edgexfoundry/edgex-go/internal/core/command"

	"github.com/labstack/echo/v4"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	command.Main(ctx, cancel, echo.New())
}
