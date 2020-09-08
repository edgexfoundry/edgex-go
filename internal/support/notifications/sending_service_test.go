/*******************************************************************************
 * Copyright 2020 Technotects
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
package notifications

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBuildSmtpMessageNoContentType(t *testing.T) {
	subject := uuid.New().String()
	message := uuid.New().String()

	result := buildSmtpMessage(subject, "", message)

	require.NotNil(t, result)

	stringResult := string(result)

	expected := fmt.Sprintf("Subject: %s\r\n\r\n%s\r\n", subject, message)
	assert.Equal(t, expected, stringResult)
}

func TestBuildSmtpMessageContentType(t *testing.T) {
	subject := uuid.New().String()
	contentType := uuid.New().String()
	message := uuid.New().String()

	result := buildSmtpMessage(subject, contentType, message)

	require.NotNil(t, result)

	stringResult := string(result)

	expected := fmt.Sprintf("Subject: %s\r\nMIME-version: 1.0;\r\nContent-Type: %s; charset=\"UTF-8\";\r\n\r\n%s\r\n", subject, contentType, message)
	assert.Equal(t, expected, stringResult)
}

func TestBuildSmtpMessageLongMessageIsChunkedIfNeeded(t *testing.T) {
	subject := uuid.New().String()
	message := uuid.New().String()

	for i := 0; i < 5; i++ {
		message += message
	}

	require.Greater(t, len(message), 998)
	require.Less(t, len(message), 1896)

	result := buildSmtpMessage(subject, "", message)

	require.NotNil(t, result)

	stringResult := string(result)

	expected := fmt.Sprintf("Subject: %s\r\n\r\n%s\r\n%s\r\n", subject, message[0:998], message[998:])
	assert.Equal(t, expected, stringResult)
}

func TestBuildSmtpMessageLongMessageIsPreChunked(t *testing.T) {
	subject := uuid.New().String()
	longLine := uuid.New().String()

	for i := 0; i < 5; i++ {
		longLine += longLine
	}

	require.Greater(t, len(longLine), 998)
	require.Less(t, len(longLine), 1896)

	formattedMessage := fmt.Sprintf("%s\r\n%s", longLine[0:998], longLine[998:])

	result := buildSmtpMessage(subject, "", formattedMessage)

	require.NotNil(t, result)

	stringResult := string(result)

	expected := fmt.Sprintf("Subject: %s\r\n\r\n%s\r\n", subject, formattedMessage)
	assert.Equal(t, expected, stringResult)
}

func TestBuildSmtpMessageLongMessageIsPartlyChunked(t *testing.T) {
	subject := uuid.New().String()
	longLine := uuid.New().String()

	for i := 0; i < 5; i++ {
		longLine += longLine
	}

	require.Greater(t, len(longLine), 998)
	require.Less(t, len(longLine), 1896)

	goodLine := uuid.New().String() + uuid.New().String() + "\r\n"

	formattedMessage := fmt.Sprintf("%s\r\n%s", longLine[0:998], longLine[998:])

	result := buildSmtpMessage(subject, "", goodLine+formattedMessage)

	require.NotNil(t, result)

	stringResult := string(result)

	expected := fmt.Sprintf("Subject: %s\r\n\r\n%s%s\r\n%s\r\n", subject, goodLine, longLine[0:998], longLine[998:])
	assert.Equal(t, expected, stringResult)
}
