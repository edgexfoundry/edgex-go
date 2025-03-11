//go:build !no_dto_validator

//
// Copyright (C) 2020-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

var val *validator.Validate

const (
	dtoDurationTag                     = "edgex-dto-duration"
	dtoUuidTag                         = "edgex-dto-uuid"
	dtoNoneEmptyStringTag              = "edgex-dto-none-empty-string"
	dtoValueType                       = "edgex-dto-value-type"
	dtoRFC3986UnreservedCharTag        = "edgex-dto-rfc3986-unreserved-chars"
	emptyOrDtoRFC3986UnreservedCharTag = "len=0|" + dtoRFC3986UnreservedCharTag
	dtoInterDatetimeTag                = "edgex-dto-interval-datetime"
	dtoNoReservedCharTag               = "edgex-dto-no-reserved-chars"
	emptyOrDtoNoReservedCharTag        = "len=0|" + dtoNoReservedCharTag
	dtoUsernameTag                     = "edgex-dto-username"
	dtoPasswordTag                     = "edgex-dto-password" // nolint:gosec
)

const (
	// Per https://tools.ietf.org/html/rfc3986#section-2.3, unreserved characters= ALPHA / DIGIT / "-" / "." / "_" / "~"
	// Also due to names used in topics for Redis Pub/Sub, "."are not allowed
	rFC3986UnreservedCharsRegexString = "^[a-zA-Z0-9-_~:;=]+$"
	intervalDatetimeLayout            = "20060102T150405"
	name                              = "Name"
	reservedCharsRegexString          = "^[^/#+$]+$"
	// Username must start and end with a letter or digit
	// The middle part can be letters, digits, underscores, or dots, with a length of 1 to 18 characters
	// The total length must be between 3 and 20 characters long
	usernameRegexString = "^[a-zA-Z0-9][a-zA-Z0-9._]{1,18}[a-zA-Z0-9]$"
)

var (
	rFC3986UnreservedCharsRegex = regexp.MustCompile(rFC3986UnreservedCharsRegexString)
	reservedCharsRegex          = regexp.MustCompile(reservedCharsRegexString)
	usernameRegex               = regexp.MustCompile(usernameRegexString)
)

func init() {
	val = validator.New()
	_ = val.RegisterValidation(dtoDurationTag, ValidateDuration)
	_ = val.RegisterValidation(dtoUuidTag, ValidateDtoUuid)
	_ = val.RegisterValidation(dtoNoneEmptyStringTag, ValidateDtoNoneEmptyString)
	_ = val.RegisterValidation(dtoValueType, ValidateValueType)
	_ = val.RegisterValidation(dtoRFC3986UnreservedCharTag, ValidateDtoRFC3986UnreservedChars)
	_ = val.RegisterValidation(dtoInterDatetimeTag, ValidateIntervalDatetime)
	_ = val.RegisterValidation(dtoNoReservedCharTag, ValidateDtoNoReservedChars)
	_ = val.RegisterValidation(dtoUsernameTag, ValidateDtoUsername)
	_ = val.RegisterValidation(dtoPasswordTag, ValidateDtoPassword)
}

// Validate function will use the validator package to validate the struct annotation
func Validate(a interface{}) error {
	err := val.Struct(a)
	// translate all error at once
	if err != nil {
		errs := err.(validator.ValidationErrors)
		var errMsg []string
		for _, e := range errs {
			errMsg = append(errMsg, getErrorMessage(e))
		}
		return errors.NewCommonEdgeX(errors.KindContractInvalid, strings.Join(errMsg, "; "), nil)
	}
	return nil
}

// Internal: generate representative validation error messages
func getErrorMessage(e validator.FieldError) string {
	tag := e.Tag()
	// StructNamespace returns the namespace for the field error, with the field's actual name.
	fieldName := e.StructNamespace()
	fieldValue := e.Param()
	var msg string
	switch tag {
	case "uuid":
		msg = fmt.Sprintf("%s field needs a uuid", fieldName)
	case "required":
		msg = fmt.Sprintf("%s field is required", fieldName)
	case "required_without":
		msg = fmt.Sprintf("%s field is required if the %s is not present", fieldName, fieldValue)
	case "len":
		msg = fmt.Sprintf("The length of %s field is not %s", fieldName, fieldValue)
	case "oneof":
		msg = fmt.Sprintf("%s field should be one of %s", fieldName, fieldValue)
	case "gt":
		msg = fmt.Sprintf("%s field should greater than %s", fieldName, fieldValue)
	case "unique":
		// If the tag contains a field param, it means that the unique tag is used for a slice of struct
		if fieldValue != "" {
			msg = fmt.Sprintf("%s field should only contain unique elements with unique '%s' values", fieldName, fieldValue)
		} else {
			msg = fmt.Sprintf("%s field should only contain unique elements", fieldName)
		}
	case dtoDurationTag:
		msg = fmt.Sprintf("%s field should follows the ISO 8601 Durations format, e.g.,100ms, 24h, or be greater than or equal to the minimum value %s ", fieldName, fieldValue)
	case dtoUuidTag:
		msg = fmt.Sprintf("%s field needs a uuid", fieldName)
	case dtoNoneEmptyStringTag:
		msg = fmt.Sprintf("%s field should not be empty string", fieldName)
	case dtoRFC3986UnreservedCharTag, emptyOrDtoRFC3986UnreservedCharTag:
		msg = fmt.Sprintf("%s field only allows unreserved characters which are ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_~:;=", fieldName)
	case dtoNoReservedCharTag, emptyOrDtoNoReservedCharTag:
		msg = fmt.Sprintf("%s field does not allow reserved characters which are /#+$", fieldName)
	case dtoUsernameTag:
		msg = fmt.Sprintf("%s field must start and end with a letter or digit. The middle part allows letters, digits, underscores, or dots. The total length must be between 3 and 20 characters long.", fieldName)
	case dtoPasswordTag:
		msg = fmt.Sprintf("%s field must contain at least 1 uppercase and lowercase letters, 1 digit, and 1 special character @$!%%*?&. The total length must be between 8 and 64 characters long.", fieldName)
	default:
		msg = fmt.Sprintf("%s field validation failed on the %s tag with value '%s'", fieldName, tag, fieldValue)
	}
	return msg
}

// ValidateDuration validate field which should follow the ISO 8601 Durations format
// the min/max of the Duration can be set via the tag params
// ex. edgex-dto-duration=10ms0x2C24h - 10ms represents the minimum Duration and 24h represents the maximum Duration
// 0x2c is the UTF-8 hex encoding of comma (,) as the min/max value separator
func ValidateDuration(fl validator.FieldLevel) bool {
	duration, err := time.ParseDuration(fl.Field().String())
	if err != nil {
		return false
	}

	// if min/max are defined from tag param, check if the duration value is in the duration range
	param := fl.Param()
	var min, max time.Duration
	if param != "" {
		params := strings.Split(param, CommaSeparator)
		if len(params) > 0 {
			min, err = time.ParseDuration(params[0])
			if err != nil {
				return false
			}
			if duration < min {
				// the duration value is smaller than the min
				return false
			}
			if len(params) > 1 {
				max, err = time.ParseDuration(params[1])
				if err != nil {
					return false
				}
				if duration > max {
					// the duration value is larger than the max
					return false
				}
			}
		}
	}
	return true
}

// ValidateDtoUuid used to check the UpdateDTO uuid pointer value
// Currently, required_without can not correct work with other tag, so write custom tag instead.
// Issue can refer to https://github.com/go-playground/validator/issues/624
func ValidateDtoUuid(fl validator.FieldLevel) bool {
	idField := fl.Field()
	// Skip the validation if the pointer value is nil
	if isNilPointer(idField) {
		return true
	}

	// The Id field should accept the empty string if the Name field is provided
	nameField := fl.Parent().FieldByName(name)
	if len(strings.TrimSpace(idField.String())) == 0 && !isNilPointer(nameField) && len(nameField.Elem().String()) > 0 {
		return true
	}

	_, err := uuid.Parse(idField.String())
	return err == nil
}

// ValidateDtoNoneEmptyString used to check the UpdateDTO name pointer value
func ValidateDtoNoneEmptyString(fl validator.FieldLevel) bool {
	val := fl.Field()
	// Skip the validation if the pointer value is nil
	if isNilPointer(val) {
		return true
	}
	// The string value should not be empty
	if len(strings.TrimSpace(val.String())) > 0 {
		return true
	} else {
		return false
	}
}

// ValidateValueType checks whether the valueType is valid
func ValidateValueType(fl validator.FieldLevel) bool {
	valueType := fl.Field().String()
	for _, v := range valueTypes {
		if strings.EqualFold(valueType, v) {
			return true
		}
	}
	return false
}

// ValidateDtoRFC3986UnreservedChars used to check if DTO's name pointer value only contains unreserved characters as
// defined in https://tools.ietf.org/html/rfc3986#section-2.3
func ValidateDtoRFC3986UnreservedChars(fl validator.FieldLevel) bool {
	val := fl.Field()
	// Skip the validation if the pointer value is nil
	if isNilPointer(val) {
		return true
	} else {
		return rFC3986UnreservedCharsRegex.MatchString(val.String())
	}
}

// ValidateIntervalDatetime validate Interval's datetime field which should follow the ISO 8601 format YYYYMMDD'T'HHmmss
func ValidateIntervalDatetime(fl validator.FieldLevel) bool {
	_, err := time.Parse(intervalDatetimeLayout, fl.Field().String())
	return err == nil
}

func isNilPointer(value reflect.Value) bool {
	return value.Kind() == reflect.Ptr && value.IsNil()
}

// ValidateDtoNoReservedChars used to check if DTO's name pointer value excludes reserved characters= / "/" / "#" / "." / "*" / "+" / "$"
func ValidateDtoNoReservedChars(fl validator.FieldLevel) bool {
	val := fl.Field()
	// Skip the validation if the pointer value is nil
	if isNilPointer(val) {
		return true
	} else {
		return reservedCharsRegex.MatchString(val.String())
	}
}

// ValidateDtoUsername used to check if DTO's username field follows the usernameRegex rule
func ValidateDtoUsername(fl validator.FieldLevel) bool {
	val := fl.Field()
	// Skip the validation if the pointer value is nil
	if isNilPointer(val) {
		return true
	} else {
		return usernameRegex.MatchString(val.String())
	}
}

// ValidateDtoPassword used to check if DTO's password field contains at least 1 uppercase letter, 1 lowercase letter, 1 digit
// and 1 special character (one of @$!%*?&); the password length is 8 to 64 characters long
func ValidateDtoPassword(fl validator.FieldLevel) bool {
	val := fl.Field()
	// Skip the validation if the pointer value is nil
	if isNilPointer(val) {
		return true
	}

	password := val.String()
	// Password length should be in the range of 8-64 characters
	if len(password) < 8 || len(password) > 64 {
		return false
	}

	// Check if the password contains at least 1 uppercase letter, 1 lowercase letter, 1 digit, and 1 special character (one of @$!%*?&)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString
	hasSpecialChar := regexp.MustCompile(`[@$!%*?&]`).MatchString

	return hasLower(password) && hasUpper(password) && hasNumber(password) && hasSpecialChar(password)
}
