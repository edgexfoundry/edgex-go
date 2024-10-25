/*******************************************************************************
 * Copyright 2021 Intel Corporation
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
 *
 *******************************************************************************/

package tokenfilewriter

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-secrets/v4/pkg/token/fileioperformer"
	"github.com/edgexfoundry/go-mod-secrets/v4/secrets/mocks"
)

var lc logger.LoggingClient
var flOpener fileioperformer.FileIoPerformer

func TestMain(m *testing.M) {
	lc = logger.MockLogger{}
	flOpener = fileioperformer.NewDefaultFileIoPerformer()
	os.Exit(m.Run())
}

func TestNewTokenFileWriter(t *testing.T) {
	sc := &mocks.SecretStoreClient{}
	tokenFileWriter := NewWriter(lc, sc, flOpener)
	require.NotEmpty(t, tokenFileWriter)
}
