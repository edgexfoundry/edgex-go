package errorz

import "net/http"

const (
	NotFoundCode    string = "NOT_FOUND"
	NotFoundMessage string = "The resource requested was not found or is no longer available"
	NotFoundStatus  int    = http.StatusNotFound

	UnhandledCode    string = "UNHANDLED"
	UnhandledMessage string = "An unhandled error occurred"
	UnhandledStatus  int    = http.StatusInternalServerError

	InvalidFieldCode    string = "INVALID_FIELD"
	InvalidFieldMessage string = "The field contains an invalid value"
	InvalidFieldStatus  int    = http.StatusBadRequest

	EntityCanNotBeDeletedCode    string = "ENTITY_CAN_NOT_BE_DELETED"
	EntityCanNotBeDeletedMessage string = "The entity requested for delete can not be deleted"
	EntityCanNotBeDeletedStatus         = http.StatusBadRequest

	EntityCanNotBeUpdatedCode    string = "ENTITY_CAN_NOT_BE_UPDATED"
	EntityCanNotBeUpdatedMessage string = "The entity requested for update can not be updated"
	EntityCanNotBeUpdatedStatus         = http.StatusBadRequest

	CouldNotValidateCode    string = "COULD_NOT_VALIDATE"
	CouldNotValidateMessage string = "The supplied request contains an invalid document or no valid accept content were available, see cause"
	CouldNotValidateStatus  int    = http.StatusBadRequest

	UnauthorizedCode    string = "UNAUTHORIZED"
	UnauthorizedMessage string = "The request could not be completed. The session is not authorized or the credentials are invalid"
	UnauthorizedStatus  int    = http.StatusUnauthorized
)

// specific
const (
	InvalidFilterCode       string = "INVALID_FILTER"
	InvalidFilterMessage    string = "The filter query supplied is invalid"
	httpStatusInvalidFilter        = http.StatusBadRequest
	InvalidFilterStatus     int    = httpStatusInvalidFilter

	InvalidPaginationCode    string = "INVALID_PAGINATION"
	InvalidPaginationMessage string = "The pagination properties provided are invalid"
	InvalidPaginationStatus  int    = http.StatusBadRequest

	InvalidSortCode    string = "INVALID_SORT_IDENTIFIER"
	InvalidSortMessage string = "The sort order supplied is invalid"
	InvalidSortStatus  int    = http.StatusBadRequest
)
