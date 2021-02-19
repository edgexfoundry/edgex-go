//
// Copyright (c) 2021 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package secretstore

import (
	"errors"
	"testing"

	. "github.com/edgexfoundry/edgex-go/internal/security/kdf/mocks"
	. "github.com/edgexfoundry/edgex-go/internal/security/pipedhexreader/mocks"
	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/types"

	"github.com/edgexfoundry/go-mod-secrets/v2/pkg/token/fileioperformer/mocks"

	"github.com/stretchr/testify/require"
)

// TestVMKEncryptionNoIkm tests the no-op path
func TestVMKEncryptionNoIkm(t *testing.T) {
	// Arrange
	fileOpener := &mocks.FileIoPerformer{}
	pipedHexReader := &MockPipedHexReader{}
	kdf := &MockKeyDeriver{}

	// Act / Assert
	vmkEncryption := NewVMKEncryption(fileOpener, pipedHexReader, kdf)
	encrypting := vmkEncryption.IsEncrypting()
	require.False(t, encrypting)

	vmkEncryption.WipeIKM()

	fileOpener.AssertExpectations(t)
	pipedHexReader.AssertExpectations(t)
	kdf.AssertExpectations(t)
}

// TestVMKEncryption tests the happy path
func TestVMKEncryption(t *testing.T) {
	// Arrange
	fakeIkm := make([]byte, 512)
	fileOpener := &mocks.FileIoPerformer{}
	pipedHexReader := &MockPipedHexReader{}
	pipedHexReader.On("ReadHexBytesFromExe", "/bin/myikm").Return(fakeIkm, nil)
	kdf := &MockKeyDeriver{}
	kdf.On("DeriveKey", make([]byte, 512), uint(32), "vault0").Return(make([]byte, 32), nil)
	kdf.On("DeriveKey", make([]byte, 512), uint(32), "vault1").Return(make([]byte, 32), nil)
	initialInitResp := types.InitResponse{
		Keys:       []string{"aabbcc", "ddeeff"},
		KeysBase64: []string{"qrvM", "3e7/"},
	}
	initResp := initialInitResp

	// Act & Assert
	vmkEncryption := NewVMKEncryption(fileOpener, pipedHexReader, kdf)
	err := vmkEncryption.LoadIKM("/bin/myikm")
	require.NoError(t, err)

	err = vmkEncryption.EncryptInitResponse(&initResp)
	require.NoError(t, err)

	err = vmkEncryption.DecryptInitResponse(&initResp)
	require.NoError(t, err)
	require.Equal(t, initialInitResp, initResp)

	vmkEncryption.WipeIKM()

	fileOpener.AssertExpectations(t)
	pipedHexReader.AssertExpectations(t)
	kdf.AssertExpectations(t)
}

// TestVMKEncryptionFailPath tests the fail path
func TestVMKEncryptionFailPath(t *testing.T) {
	// Arrange
	fakeIkm := make([]byte, 512)
	fileOpener := &mocks.FileIoPerformer{}
	pipedHexReader := &MockPipedHexReader{}
	pipedHexReader.On("ReadHexBytesFromExe", "/bin/myikm").Return(fakeIkm, errors.New("error"))
	kdf := &MockKeyDeriver{}
	initialInitResp := types.InitResponse{
		Keys: []string{"aabbcc", "ddeeff"},
	}
	initResp := initialInitResp

	// Act & Assert
	vmkEncryption := NewVMKEncryption(fileOpener, pipedHexReader, kdf)
	err := vmkEncryption.LoadIKM("/bin/myikm")
	require.Error(t, err)

	err = vmkEncryption.EncryptInitResponse(&initResp)
	require.Error(t, err)

	err = vmkEncryption.DecryptInitResponse(&initResp)
	require.Error(t, err)

	vmkEncryption.WipeIKM()

	fileOpener.AssertExpectations(t)
	pipedHexReader.AssertExpectations(t)
	kdf.AssertExpectations(t)
}
