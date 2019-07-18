//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//
package option

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

func TestNewPkiInitOption_GenerateOnly(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		GenerateOpt: true,
	}
	// generate option given
	generateOn, _, _ := NewPkiInitOption(options)
	assert.NotNil(generateOn)
	assert.Equal(true, generateOn.(*PkiInitOption).GenerateOpt)

	// generate option omitted
	options.GenerateOpt = false
	generateOff, _, _ := NewPkiInitOption(options)
	assert.NotNil(generateOff)
	assert.Equal(false, generateOff.(*PkiInitOption).GenerateOpt)
}

func TestNewPkiInitOption_ImportOnly(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		ImportOpt: true,
	}
	// import option given
	optionsExecutor, _, _ := NewPkiInitOption(options)
	assert.NotNil(optionsExecutor)
	assert.Equal(true, optionsExecutor.(*PkiInitOption).ImportOpt)

	// import option omitted
	options.ImportOpt = false
	optionsExecutor, _, _ = NewPkiInitOption(options)
	assert.NotNil(optionsExecutor)
	assert.Equal(false, optionsExecutor.(*PkiInitOption).ImportOpt)
}

func TestNewPkiInitOption_CacheOnly(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		CacheOpt: true,
	}
	// cache option given
	optionsExecutor, _, _ := NewPkiInitOption(options)
	assert.NotNil(optionsExecutor)
	assert.Equal(true, optionsExecutor.(*PkiInitOption).CacheOpt)

	// cache option omitted
	options.CacheOpt = false
	optionsExecutor, _, _ = NewPkiInitOption(options)
	assert.NotNil(optionsExecutor)
	assert.Equal(false, optionsExecutor.(*PkiInitOption).CacheOpt)
}

func TestNewPkiInitOption_CacheCAOnly(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		CacheCAOpt: true,
	}
	// cache CA option given
	optionsExecutor, _, _ := NewPkiInitOption(options)
	assert.NotNil(optionsExecutor)
	assert.Equal(true, optionsExecutor.(*PkiInitOption).CacheCAOpt)

	// cache CA option omitted
	options.CacheCAOpt = false
	optionsExecutor, _, _ = NewPkiInitOption(options)
	assert.NotNil(optionsExecutor)
	assert.Equal(false, optionsExecutor.(*PkiInitOption).CacheCAOpt)
}

func TestNewPkiInitOption_ImportAndGenerate(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		GenerateOpt: true,
		ImportOpt:   true,
	}
	// import and generate option given
	optionsExecutor, status, err := NewPkiInitOption(options)
	assert.Empty(optionsExecutor)
	assert.Equal(exitWithError.intValue(), status)
	assert.NotNil(err)
}

func TestNewPkiInitOption_CacheAndGenerate(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		GenerateOpt: true,
		CacheOpt:    true,
	}
	optionsExecutor, status, err := NewPkiInitOption(options)
	assert.Empty(optionsExecutor)
	assert.Equal(exitWithError.intValue(), status)
	assert.NotNil(err)
}

func TestNewPkiInitOption_CacheAndImport(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		ImportOpt: true,
		CacheOpt:  true,
	}
	optionsExecutor, status, err := NewPkiInitOption(options)
	assert.Empty(optionsExecutor)
	assert.Equal(exitWithError.intValue(), status)
	assert.NotNil(err)
}

func TestNewPkiInitOption_CacheAndGenerateAndImport(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		GenerateOpt: true,
		ImportOpt:   true,
		CacheOpt:    true,
	}
	optionsExecutor, status, err := NewPkiInitOption(options)
	assert.Empty(optionsExecutor)
	assert.Equal(exitWithError.intValue(), status)
	assert.NotNil(err)
}

func TestNewPkiInitOption_CacheCAAndGenerate(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		GenerateOpt: true,
		CacheCAOpt:  true,
	}
	optionsExecutor, status, err := NewPkiInitOption(options)
	assert.Empty(optionsExecutor)
	assert.Equal(exitWithError.intValue(), status)
	assert.NotNil(err)
}

func TestNewPkiInitOption_CacheCAAndImport(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		ImportOpt:  true,
		CacheCAOpt: true,
	}
	optionsExecutor, status, err := NewPkiInitOption(options)
	assert.Empty(optionsExecutor)
	assert.Equal(exitWithError.intValue(), status)
	assert.NotNil(err)
}

func TestNewPkiInitOption_CacheCAAndCache(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		CacheOpt:   true,
		CacheCAOpt: true,
	}
	optionsExecutor, status, err := NewPkiInitOption(options)
	assert.Empty(optionsExecutor)
	assert.Equal(exitWithError.intValue(), status)
	assert.NotNil(err)
}

func TestNewPkiInitOption_CacheCAAndGenerateAndImport(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		GenerateOpt: true,
		ImportOpt:   true,
		CacheCAOpt:  true,
	}
	optionsExecutor, status, err := NewPkiInitOption(options)
	assert.Empty(optionsExecutor)
	assert.Equal(exitWithError.intValue(), status)
	assert.NotNil(err)
}

func TestNewPkiInitOption_CacheCAAndGenerateAndCache(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		GenerateOpt: true,
		CacheOpt:    true,
		CacheCAOpt:  true,
	}
	optionsExecutor, status, err := NewPkiInitOption(options)
	assert.Empty(optionsExecutor)
	assert.Equal(exitWithError.intValue(), status)
	assert.NotNil(err)
}

func TestNewPkiInitOption_CacheCAAndImportAndCache(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		ImportOpt:  true,
		CacheOpt:   true,
		CacheCAOpt: true,
	}
	optionsExecutor, status, err := NewPkiInitOption(options)
	assert.Empty(optionsExecutor)
	assert.Equal(exitWithError.intValue(), status)
	assert.NotNil(err)
}

func TestNewPkiInitOption_CacheCAAndGenerateImportAndCache(t *testing.T) {
	assert := assert.New(t)

	options := PkiInitOption{
		GenerateOpt: true,
		ImportOpt:   true,
		CacheOpt:    true,
		CacheCAOpt:  true,
	}
	optionsExecutor, status, err := NewPkiInitOption(options)
	assert.Empty(optionsExecutor)
	assert.Equal(exitWithError.intValue(), status)
	assert.NotNil(err)
}

func getMockArguments() []interface{} {
	var ifc []interface{}
	// use reflection to find out how many bool type of options in PkiInitOption struct
	// and create a slice of mock argument interface instances
	elm := reflect.ValueOf(&PkiInitOption{}).Elem()
	for i := 0; i < elm.NumField(); i++ {
		field := elm.Field(i)
		switch field.Kind() {
		case reflect.Bool:
			ifc = append(ifc, mock.AnythingOfTypeArgument("func(*option.PkiInitOption) (option.exitCode, error)"))
		}
	}

	return ifc
}
func TestProcessOptionNormal(t *testing.T) {
	testExecutor := &mockOptionsExecutor{}
	// normal case
	testExecutor.On("executeOptions", getMockArguments()...).
		Return(normal, nil).Once()

	assert := assert.New(t)

	options := PkiInitOption{
		GenerateOpt: true,
		ImportOpt:   false,
	}
	optsExecutor, _, _ := NewPkiInitOption(options)
	optsExecutor.(*PkiInitOption).executor = testExecutor
	exitCode, err := optsExecutor.ProcessOptions()
	assert.Equal(normal.intValue(), exitCode)
	assert.Nil(err)

	testExecutor.AssertExpectations(t)
}

func TestProcessOptionError(t *testing.T) {
	testExecutor := &mockOptionsExecutor{}
	generateErr := errors.New("failed to execute generate")
	// error case
	testExecutor.On("executeOptions", getMockArguments()...).
		Return(exitWithError, generateErr).Once()

	assert := assert.New(t)

	options := PkiInitOption{
		GenerateOpt: true,
		ImportOpt:   false,
	}
	generateOn, _, _ := NewPkiInitOption(options)
	generateOn.(*PkiInitOption).executor = testExecutor
	exitCode, err := generateOn.ProcessOptions()
	assert.Equal(exitWithError.intValue(), exitCode)
	assert.Equal(generateErr, err)

	testExecutor.AssertExpectations(t)
}

func TestExecuteOption(t *testing.T) {
	testExecutor := &mockOptionsExecutor{}
	assert := assert.New(t)

	options := PkiInitOption{
		GenerateOpt: true,
	}
	generateOn, _, _ := NewPkiInitOption(options)
	generateOn.(*PkiInitOption).executor = testExecutor
	exitCode, err := generateOn.executeOptions(mockGenerate())
	assert.Equal(normal, exitCode)
	assert.Nil(err)

	options.GenerateOpt = false
	generateOff, _, _ := NewPkiInitOption(options)
	generateOff.(*PkiInitOption).executor = testExecutor
	exitCode, err = generateOff.executeOptions(mockGenerate())
	assert.Equal(normal, exitCode)
	assert.Nil(err)
}

func mockGenerate() func(*PkiInitOption) (exitCode, error) {
	return func(pkiInitOpton *PkiInitOption) (exitCode, error) {
		return normal, nil
	}
}
