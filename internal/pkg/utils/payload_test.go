package utils

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckPayloadSize(t *testing.T) {
	smallpayload := make([]byte, 10)
	largePayload := make([]byte, 25000*1024)
	tests := []struct {
		name            string
		payload         []byte
		sizeLimit       int64
		errorExpected   bool
		expectedErrKind errors.ErrKind
	}{
		{"Valid small size", smallpayload, int64(len(smallpayload)), false, ""},
		{"Valid large size", largePayload, int64(len(largePayload)), false, ""},
		{"Invalid small size", smallpayload, int64(len(smallpayload) - 1), true, errors.KindLimitExceeded},
		{"Invalid large size", largePayload, int64(len(largePayload) - 1), true, errors.KindLimitExceeded},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := CheckPayloadSize(testCase.payload, testCase.sizeLimit)
			if testCase.errorExpected {
				require.Error(t, err)
				assert.NotEmpty(t, err.Error(), "Error message is empty")
				assert.Equal(t, testCase.expectedErrKind, errors.Kind(err), "Error kind not as expected")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
