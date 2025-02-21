//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common_config

import (
	"os"
	"testing"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"

	"github.com/stretchr/testify/require"
)

func TestGetOverwriteConfig(t *testing.T) {
	lc := logger.NewMockClient()

	tests := []struct {
		name           string
		flags          []string
		hasEnvVar      bool
		envValue       string
		expectedResult bool
	}{
		{"Overwrite config with -o flag", []string{"-o"}, false, "", true},
		{"Overwrite config with EDGEX_OVERWRITE_CONFIG env variable is true", []string{}, true, "true", true},
		{"Not overwrite config with -o flag and EDGEX_OVERWRITE_CONFIG env variable is false", []string{"-o"}, true, "false", false},
		{"Not overwrite config no -o flag and no EDGEX_OVERWRITE_CONFIG env variable defined", []string{}, false, "", false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			f := flags.New()
			args := testCase.flags
			f.Parse(args)

			if testCase.hasEnvVar {
				err := os.Setenv("EDGEX_OVERWRITE_CONFIG", testCase.envValue)
				require.NoError(t, err)
			}

			overwriteConfig := getOverwriteConfig(f, lc)
			require.Equal(t, testCase.expectedResult, overwriteConfig)

			err := os.Unsetenv("EDGEX_OVERWRITE_CONFIG")
			require.NoError(t, err)
		})
	}
}
