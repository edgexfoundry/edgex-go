//
// Copyright (c) 2020 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package kdf

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-secrets/v3/pkg/token/fileioperformer/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const expectedKey = "1060e4e72054653bf46623844033f5ccc9cff596a4a680e074ef4fd06aae60df"

// TestNoErrorKdfCreateSalt tests the golden path for a new salt
func TestNoErrorKdfCreateSalt(t *testing.T) {
	// Arrange
	mockFileInfo := &mockFileInfo{}
	defer mockOsStat(func(string) (os.FileInfo, error) { return mockFileInfo, os.ErrNotExist })()
	mockSeedFile := &mockSeedFile{}
	mockSeedFile.On("Write", mock.Anything).Return(32, nil)
	mockSeedFile.On("Close").Return(nil)
	mockFileOpener := &mocks.FileIoPerformer{}
	mockFileOpener.On("OpenFileWriter", "/target/kdf-salt.dat", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(0600)).Return(mockSeedFile, nil)
	keyDeriver := NewKdf(mockFileOpener, "/target", sha256.New)

	// Act
	_, err := keyDeriver.DeriveKey(make([]byte, 32), 32, "info")

	// Assert
	mockFileOpener.AssertExpectations(t)
	mockSeedFile.AssertExpectations(t)
	require.NoError(t, err)
	// Output key expected to be random; unable to assert
}

// TestNoErrorKdfReadSalt tests the golden path reading a salt
func TestNoErrorKdfReadSalt(t *testing.T) {
	// Arrange
	mockFileInfo := &mockFileInfo{}
	defer mockOsStat(func(string) (os.FileInfo, error) { return mockFileInfo, nil })()
	mockSeedFile := &mockSeedFile{}
	mockSeedFile.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		b := args.Get(0).([]byte)
		for i := range b {
			b[i] = 0
		}
	}).Return(32, nil)
	mockSeedFile.On("Close").Return(nil)
	mockFileOpener := &mocks.FileIoPerformer{}
	mockFileOpener.On("OpenFileReader", "/target/kdf-salt.dat", os.O_RDONLY, os.FileMode(0400)).Return(mockSeedFile, nil)
	keyDeriver := NewKdf(mockFileOpener, "/target", sha256.New)
	expected, _ := hex.DecodeString(expectedKey)

	// Act
	key, err := keyDeriver.DeriveKey(make([]byte, 32), 32, "info")

	// Assert
	mockFileOpener.AssertExpectations(t)
	mockSeedFile.AssertExpectations(t)
	require.NoError(t, err)
	require.Equal(t, expected, key)
}

// TestFailedStat tests os.Stat() returning unexpected value
func TestFailedStat(t *testing.T) {
	// Arrange
	mockFileInfo := &mockFileInfo{}
	defer mockOsStat(func(string) (os.FileInfo, error) { return mockFileInfo, os.ErrPermission })()
	mockSeedFile := &mockSeedFile{}
	mockFileOpener := &mocks.FileIoPerformer{}
	keyDeriver := NewKdf(mockFileOpener, "/target", sha256.New)

	// Act
	key, err := keyDeriver.DeriveKey(make([]byte, 32), 32, "info")

	// Assert
	mockFileOpener.AssertExpectations(t)
	mockSeedFile.AssertExpectations(t)
	require.Error(t, err)
	require.Nil(t, key)
}

func TestFailedFileOpenForReading(t *testing.T) {
	// Arrange
	mockFileInfo := &mockFileInfo{}
	defer mockOsStat(func(string) (os.FileInfo, error) { return mockFileInfo, nil })()
	mockSeedFile := &mockSeedFile{}
	mockFileOpener := &mocks.FileIoPerformer{}
	mockFileOpener.On("OpenFileReader", "/target/kdf-salt.dat", os.O_RDONLY, os.FileMode(0400)).Return(mockSeedFile, errors.New("error"))
	keyDeriver := NewKdf(mockFileOpener, "/target", sha256.New)

	// Act
	_, err := keyDeriver.DeriveKey(make([]byte, 32), 32, "info")

	// Assert
	mockFileOpener.AssertExpectations(t)
	mockSeedFile.AssertExpectations(t)
	require.Error(t, err)
}

func TestFailedRead(t *testing.T) {
	// Arrange
	mockFileInfo := &mockFileInfo{}
	defer mockOsStat(func(string) (os.FileInfo, error) { return mockFileInfo, nil })()
	mockSeedFile := &mockSeedFile{}
	mockSeedFile.On("Read", mock.Anything).Return(0, errors.New("error"))
	mockSeedFile.On("Close").Return(nil)
	mockFileOpener := &mocks.FileIoPerformer{}
	mockFileOpener.On("OpenFileReader", "/target/kdf-salt.dat", os.O_RDONLY, os.FileMode(0400)).Return(mockSeedFile, nil)
	keyDeriver := NewKdf(mockFileOpener, "/target", sha256.New)

	// Act
	key, err := keyDeriver.DeriveKey(make([]byte, 32), 32, "info")

	// Assert
	mockFileOpener.AssertExpectations(t)
	mockSeedFile.AssertExpectations(t)
	require.Error(t, err)
	require.Nil(t, key)
}

func TestShortRead(t *testing.T) {
	// Arrange
	mockFileInfo := &mockFileInfo{}
	defer mockOsStat(func(string) (os.FileInfo, error) { return mockFileInfo, nil })()
	mockSeedFile := &mockSeedFile{}
	mockSeedFile.On("Read", mock.Anything).Return(1, nil)
	mockSeedFile.On("Close").Return(nil)
	mockFileOpener := &mocks.FileIoPerformer{}
	mockFileOpener.On("OpenFileReader", "/target/kdf-salt.dat", os.O_RDONLY, os.FileMode(0400)).Return(mockSeedFile, nil)
	keyDeriver := NewKdf(mockFileOpener, "/target", sha256.New)

	// Act
	key, err := keyDeriver.DeriveKey(make([]byte, 32), 32, "info")

	// Assert
	mockFileOpener.AssertExpectations(t)
	mockSeedFile.AssertExpectations(t)
	require.Error(t, err)
	require.Nil(t, key)
}

// TestFailedFileOpenForWriting tests failed file open path
func TestFailedFileOpenForWriting(t *testing.T) {
	// Arrange
	mockFileInfo := &mockFileInfo{}
	defer mockOsStat(func(string) (os.FileInfo, error) { return mockFileInfo, os.ErrNotExist })()
	mockSeedFile := &mockSeedFile{}
	mockFileOpener := &mocks.FileIoPerformer{}
	mockFileOpener.On("OpenFileWriter", "/target/kdf-salt.dat", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(0600)).Return(mockSeedFile, errors.New("error"))
	keyDeriver := NewKdf(mockFileOpener, "/target", sha256.New)

	// Act
	_, err := keyDeriver.DeriveKey(make([]byte, 32), 32, "info")

	// Assert
	mockFileOpener.AssertExpectations(t)
	mockSeedFile.AssertExpectations(t)
	require.Error(t, err)
}

func TestFailedWrite(t *testing.T) {
	// Arrange
	mockFileInfo := &mockFileInfo{}
	defer mockOsStat(func(string) (os.FileInfo, error) { return mockFileInfo, os.ErrNotExist })()
	mockSeedFile := &mockSeedFile{}
	mockSeedFile.On("Write", mock.Anything).Return(32, errors.New("error"))
	mockSeedFile.On("Close").Return(nil)
	mockFileOpener := &mocks.FileIoPerformer{}
	mockFileOpener.On("OpenFileWriter", "/target/kdf-salt.dat", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(0600)).Return(mockSeedFile, nil)
	keyDeriver := NewKdf(mockFileOpener, "/target", sha256.New)

	// Act
	key, err := keyDeriver.DeriveKey(make([]byte, 32), 32, "info")

	// Assert
	mockFileOpener.AssertExpectations(t)
	mockSeedFile.AssertExpectations(t)
	require.Error(t, err)
	require.Nil(t, key)
}

func TestShortWrite(t *testing.T) {
	// Arrange
	mockFileInfo := &mockFileInfo{}
	defer mockOsStat(func(string) (os.FileInfo, error) { return mockFileInfo, os.ErrNotExist })()
	mockSeedFile := &mockSeedFile{}
	mockSeedFile.On("Write", mock.Anything).Return(15, nil)
	mockSeedFile.On("Close").Return(nil)
	mockFileOpener := &mocks.FileIoPerformer{}
	mockFileOpener.On("OpenFileWriter", "/target/kdf-salt.dat", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(0600)).Return(mockSeedFile, nil)
	keyDeriver := NewKdf(mockFileOpener, "/target", sha256.New)

	// Act
	key, err := keyDeriver.DeriveKey(make([]byte, 32), 32, "info")

	// Assert
	mockFileOpener.AssertExpectations(t)
	mockSeedFile.AssertExpectations(t)
	require.Error(t, err)
	require.Nil(t, key)
}

//
// Mock opening and reading of the seed file
//

type mockSeedFile struct {
	mock.Mock
}

func (m *mockSeedFile) Close() error {
	arguments := m.Called()
	return arguments.Error(0)
}

func (m *mockSeedFile) Read(b []byte) (n int, err error) {
	arguments := m.Called(b)
	return arguments.Int(0), arguments.Error(1)
}

func (m *mockSeedFile) Write(p []byte) (n int, err error) {
	arguments := m.Called(p)
	return arguments.Int(0), arguments.Error(1)
}

//
// Mock object to return from os.Stat()
//

// mockOsStat is a testing hook; returns a function
// that restores the old value of osStat.
func mockOsStat(f func(string) (os.FileInfo, error)) func() {
	original := osStat
	osStat = f
	return func() {
		osStat = original
	}
}

type mockFileInfo struct {
	mock.Mock
}

func (m *mockFileInfo) Name() string {
	arguments := m.Called()
	return arguments.String(0)
}

func (m *mockFileInfo) Size() int64 {
	arguments := m.Called()
	return int64(arguments.Int(0))
}

func (m *mockFileInfo) Mode() os.FileMode {
	arguments := m.Called()
	return arguments.Get(0).(os.FileMode)
}

func (m *mockFileInfo) ModTime() time.Time {
	arguments := m.Called()
	return arguments.Get(0).(time.Time)
}

func (m *mockFileInfo) IsDir() bool {
	arguments := m.Called()
	return arguments.Bool(0)
}

func (m *mockFileInfo) Sys() interface{} {
	arguments := m.Called()
	return arguments.Get(0)
}
