/*******************************************************************************
 * Copyright 2017 Dell Inc.
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

package logger

type MockLogger struct {
}

func NewMockClient() LoggingClient {
	return MockLogger{}
}

func (lc MockLogger) SetLogLevel(loglevel string) error {
	return nil
}

func (lc MockLogger) Info(msg string, args ...interface{}) error {
	return nil
}

func (lc MockLogger) Debug(msg string, args ...interface{}) error {
	return nil
}

func (lc MockLogger) Error(msg string, args ...interface{}) error {
	return nil
}

func (lc MockLogger) Trace(msg string, args ...interface{}) error {
	return nil
}

func (lc MockLogger) Warn(msg string, args ...interface{}) error {
	return nil
}
