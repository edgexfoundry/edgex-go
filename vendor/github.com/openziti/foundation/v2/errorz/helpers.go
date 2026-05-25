package errorz

import (
	"fmt"
	"strings"
)

const (
	WwwAuthRealmPrimaryExtJwt   = "openziti-primary-ext-jwt"
	WwwAuthRealmSecondaryExtJwt = "openziti-secondary-ext-jwt"
	WwwAuthRealmOidc            = "openziti-oidc"
	WwwAuthRealmZtSession       = "zt-session"

	WwwAuthErrorMissing = "missing"
	WwwAuthErrorInvalid = "invalid"
	WwwAuthErrorExpired = "expired"

	WwwAuthErrorDescMissing = "no matching token was provided"
	WwwAuthErrorDescInvalid = "token is invalid"
	WwwAuthErrorDescExpired = "token expired"

	WwwAuthExtJwtId     = "id"
	WwwAuthExtJwtIssuer = "issuer"
)

func NewNotFound() *ApiError {
	return &ApiError{
		AppCode: NotFoundCode,
		Message: NotFoundMessage,
		Status:  NotFoundStatus,
	}
}

func NewUnhandled(cause error) *ApiError {
	return &ApiError{
		AppCode: UnhandledCode,
		Message: UnhandledMessage,
		Status:  UnhandledStatus,
		Cause:   cause,
	}
}

func NewEntityCanNotBeDeleted() *ApiError {
	return &ApiError{
		AppCode: EntityCanNotBeDeletedCode,
		Message: EntityCanNotBeDeletedMessage,
		Status:  EntityCanNotBeDeletedStatus,
	}
}

func NewEntityCanNotBeDeletedFrom(err error) *ApiError {
	return &ApiError{
		AppCode:     EntityCanNotBeDeletedCode,
		Message:     EntityCanNotBeDeletedMessage,
		Status:      EntityCanNotBeDeletedStatus,
		Cause:       err,
		AppendCause: true,
	}
}

func NewEntityCanNotBeUpdatedFrom(err error) *ApiError {
	return &ApiError{
		AppCode:     EntityCanNotBeUpdatedCode,
		Message:     EntityCanNotBeUpdatedMessage,
		Status:      EntityCanNotBeUpdatedStatus,
		Cause:       err,
		AppendCause: true,
	}
}

func NewFieldApiError(fieldError *FieldError) *ApiError {
	return &ApiError{
		AppCode:     InvalidFieldCode,
		Message:     InvalidFieldMessage,
		Status:      InvalidFieldStatus,
		Cause:       fieldError,
		AppendCause: true,
	}
}

func NewCouldNotValidate(err error) *ApiError {
	return &ApiError{
		AppCode: CouldNotValidateCode,
		Message: CouldNotValidateMessage,
		Status:  CouldNotValidateStatus,
		Cause:   err,
	}
}

// NewUnauthorized represents a generic unauthorized request that conveys no additional token status
func NewUnauthorized() *ApiError {
	return &ApiError{
		AppCode: UnauthorizedCode,
		Message: UnauthorizedMessage,
		Status:  UnauthorizedStatus,
	}
}

// NewUnauthorizedTokensMissing represents an unauthorized request due to a lack of any supported security token being provided
func NewUnauthorizedTokensMissing() *ApiError {
	return &ApiError{
		AppCode: UnauthorizedCode,
		Message: UnauthorizedMessage,
		Status:  UnauthorizedStatus,
		Headers: map[string][]string{
			//single string value separate by commas due to OpenAPI 2.0 header limitations (1 value for 1 header)
			//CSV values for www-authenticate allowed per RFCs.
			"WWW-Authenticate": {
				fmt.Sprintf(`zt-session realm="%s" error="%s" error_description="%s"`, WwwAuthRealmZtSession, WwwAuthErrorMissing, WwwAuthErrorDescMissing) + "," +
					fmt.Sprintf(`Bearer realm="%s" error="%s" error_description="%s"`, WwwAuthRealmOidc, WwwAuthErrorMissing, WwwAuthErrorDescMissing),
			},
		},
	}
}

// NewUnauthorizedOidcExpired represents an unauthorized request that the provided OIDC token has expired
func NewUnauthorizedOidcExpired() *ApiError {
	return &ApiError{
		AppCode: UnauthorizedCode,
		Message: UnauthorizedMessage,
		Status:  UnauthorizedStatus,
		Headers: map[string][]string{
			"WWW-Authenticate": {
				fmt.Sprintf(`Bearer realm="%s" error="%s" error_description="%s"`, WwwAuthRealmOidc, WwwAuthErrorExpired, WwwAuthErrorDescExpired),
			},
		},
	}
}

// NewUnauthorizedOidcInvalid represents an unauthorized request that the provided OIDC token is invalid
func NewUnauthorizedOidcInvalid() *ApiError {
	return &ApiError{
		AppCode: UnauthorizedCode,
		Message: UnauthorizedMessage,
		Status:  UnauthorizedStatus,
		Headers: map[string][]string{
			"WWW-Authenticate": {
				fmt.Sprintf(`Bearer realm="%s" error="%s" error_description="%s"`, WwwAuthRealmOidc, WwwAuthErrorInvalid, WwwAuthErrorDescInvalid),
			},
		},
	}
}

// NewUnauthorizedZtSessionInvalid represents an unauthorized request that the provided legacy (zt-session) token is invalid
func NewUnauthorizedZtSessionInvalid() *ApiError {
	return &ApiError{
		AppCode: UnauthorizedCode,
		Message: UnauthorizedMessage,
		Status:  UnauthorizedStatus,
		Headers: map[string][]string{
			"WWW-Authenticate": {
				fmt.Sprintf(`zt-session realm="%s" error="%s" error_description="%s"`, WwwAuthRealmZtSession, WwwAuthErrorInvalid, WwwAuthErrorDescInvalid),
			},
		},
	}
}

// NewUnauthorizedPrimaryExtTokenMissing represents an unauthorized primary auth request that the required a JWT token, ext-jwt-signers configuration, is missing
func NewUnauthorizedPrimaryExtTokenMissing(extJwtIds, issuers []string) *ApiError {

	extJwtIdsCsv := strings.Join(extJwtIds, "|")
	issuersCsv := strings.Join(issuers, "|")

	return &ApiError{
		AppCode: UnauthorizedCode,
		Message: UnauthorizedMessage,
		Status:  UnauthorizedStatus,
		Headers: map[string][]string{
			"WWW-Authenticate": {
				fmt.Sprintf(`Bearer realm="%s" error="%s" error_description="%s" %s="%s" %s="%s"`, WwwAuthRealmPrimaryExtJwt, WwwAuthErrorMissing, WwwAuthErrorDescMissing, WwwAuthExtJwtId, extJwtIdsCsv, WwwAuthExtJwtIssuer, issuersCsv),
			},
		},
	}
}

// NewUnauthorizedPrimaryExtTokenExpired represents an unauthorized primary auth request that required a JWT token, ext-jwt-signers configuration, is expired
func NewUnauthorizedPrimaryExtTokenExpired(extJwtIds, issuers []string) *ApiError {
	extJwtIdsCsv := strings.Join(extJwtIds, "|")
	issuersCsv := strings.Join(issuers, "|")

	return &ApiError{
		AppCode: UnauthorizedCode,
		Message: UnauthorizedMessage,
		Status:  UnauthorizedStatus,
		Headers: map[string][]string{
			"WWW-Authenticate": {
				fmt.Sprintf(`Bearer realm="%s" error="%s" error_description="%s" %s="%s" %s="%s"`, WwwAuthRealmPrimaryExtJwt, WwwAuthErrorExpired, WwwAuthErrorDescExpired, WwwAuthExtJwtId, extJwtIdsCsv, WwwAuthExtJwtIssuer, issuersCsv),
			},
		},
	}
}

// NewUnauthorizedPrimaryExtTokenInvalid represents an unauthorized primary auth request that the required additional JWT token, ext-jwt-signers configuration, is invalid
func NewUnauthorizedPrimaryExtTokenInvalid(extJwtIds, issuers []string) *ApiError {
	extJwtIdsCsv := strings.Join(extJwtIds, "|")
	issuersCsv := strings.Join(issuers, "|")

	return &ApiError{
		AppCode: UnauthorizedCode,
		Message: UnauthorizedMessage,
		Status:  UnauthorizedStatus,
		Headers: map[string][]string{
			"WWW-Authenticate": {
				fmt.Sprintf(`Bearer realm="%s" error="%s" error_description="%s" %s="%s" %s="%s"`, WwwAuthRealmPrimaryExtJwt, WwwAuthErrorInvalid, WwwAuthErrorDescInvalid, WwwAuthExtJwtId, extJwtIdsCsv, WwwAuthExtJwtIssuer, issuersCsv),
			},
		},
	}
}

//--------------

// NewUnauthorizedSecondaryExtTokenMissing represents an unauthorized request that the required additional JWT token, ext-jwt-signers configuration, is missing
func NewUnauthorizedSecondaryExtTokenMissing(extJwtIds, issuers []string) *ApiError {

	extJwtIdsCsv := strings.Join(extJwtIds, "|")
	issuersCsv := strings.Join(issuers, "|")

	return &ApiError{
		AppCode: UnauthorizedCode,
		Message: UnauthorizedMessage,
		Status:  UnauthorizedStatus,
		Headers: map[string][]string{
			"WWW-Authenticate": {
				fmt.Sprintf(`Bearer realm="%s" error="%s" error_description="%s" %s="%s" %s="%s"`, WwwAuthRealmSecondaryExtJwt, WwwAuthErrorMissing, WwwAuthErrorDescMissing, WwwAuthExtJwtId, extJwtIdsCsv, WwwAuthExtJwtIssuer, issuersCsv),
			},
		},
	}
}

// NewUnauthorizedSecondaryExtTokenExpired represents an unauthorized request that the required additional JWT token, ext-jwt-signers configuration, is expired
func NewUnauthorizedSecondaryExtTokenExpired(extJwtIds, issuers []string) *ApiError {
	extJwtIdsCsv := strings.Join(extJwtIds, "|")
	issuersCsv := strings.Join(issuers, "|")

	return &ApiError{
		AppCode: UnauthorizedCode,
		Message: UnauthorizedMessage,
		Status:  UnauthorizedStatus,
		Headers: map[string][]string{
			"WWW-Authenticate": {
				fmt.Sprintf(`Bearer realm="%s" error="%s" error_description="%s" %s="%s" %s="%s"`, WwwAuthRealmSecondaryExtJwt, WwwAuthErrorExpired, WwwAuthErrorDescExpired, WwwAuthExtJwtId, extJwtIdsCsv, WwwAuthExtJwtIssuer, issuersCsv),
			},
		},
	}
}

// NewUnauthorizedSecondaryExtTokenInvalid represents an unauthorized request that the required additional JWT token, ext-jwt-signers configuration, is invalid
func NewUnauthorizedSecondaryExtTokenInvalid(extJwtIds, issuers []string) *ApiError {
	extJwtIdsCsv := strings.Join(extJwtIds, "|")
	issuersCsv := strings.Join(issuers, "|")

	return &ApiError{
		AppCode: UnauthorizedCode,
		Message: UnauthorizedMessage,
		Status:  UnauthorizedStatus,
		Headers: map[string][]string{
			"WWW-Authenticate": {
				fmt.Sprintf(`Bearer realm="%s" error="%s" error_description="%s" %s="%s" %s="%s"`, WwwAuthRealmSecondaryExtJwt, WwwAuthErrorInvalid, WwwAuthErrorDescInvalid, WwwAuthExtJwtId, extJwtIdsCsv, WwwAuthExtJwtIssuer, issuersCsv),
			},
		},
	}
}

func NewInvalidFilter(cause error) *ApiError {
	return &ApiError{
		AppCode:     InvalidFilterCode,
		Message:     InvalidFilterMessage,
		Status:      InvalidFilterStatus,
		Cause:       cause,
		AppendCause: true,
	}
}

func NewInvalidPagination(err error) *ApiError {
	return &ApiError{
		AppCode:     InvalidPaginationCode,
		Message:     InvalidPaginationMessage,
		Status:      InvalidPaginationStatus,
		Cause:       err,
		AppendCause: true,
	}
}

func NewInvalidSort(err error) *ApiError {
	return &ApiError{
		AppCode:     InvalidSortCode,
		Message:     InvalidSortMessage,
		Status:      InvalidSortStatus,
		Cause:       err,
		AppendCause: true,
	}
}
