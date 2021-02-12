//
// Copyright (c) 2021 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package secretstore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordsAreRandom(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cg := NewDefaultCredentialGenerator()
	cred1, err := cg.Generate(ctx)
	assert.NoError(t, err)
	cred2, err := cg.Generate(ctx)
	assert.NoError(t, err)
	assert.NotEqual(t, cred1, cred2)
	defer cancel()
}
