package errorz

func NewNotFound() *ApiError {
	return &ApiError{
		Code:    NotFoundCode,
		Message: NotFoundMessage,
		Status:  NotFoundStatus,
	}
}

func NewUnhandled(cause error) *ApiError {
	return &ApiError{
		Code:    UnhandledCode,
		Message: UnhandledMessage,
		Status:  UnhandledStatus,
		Cause:   cause,
	}
}

func NewEntityCanNotBeDeleted() *ApiError {
	return &ApiError{
		Code:    EntityCanNotBeDeletedCode,
		Message: EntityCanNotBeDeletedMessage,
		Status:  EntityCanNotBeDeletedStatus,
	}
}

func NewEntityCanNotBeDeletedFrom(err error) *ApiError {
	return &ApiError{
		Code:        EntityCanNotBeDeletedCode,
		Message:     EntityCanNotBeDeletedMessage,
		Status:      EntityCanNotBeDeletedStatus,
		Cause:       err,
		AppendCause: true,
	}
}

func NewEntityCanNotBeUpdatedFrom(err error) *ApiError {
	return &ApiError{
		Code:        EntityCanNotBeUpdatedCode,
		Message:     EntityCanNotBeUpdatedMessage,
		Status:      EntityCanNotBeUpdatedStatus,
		Cause:       err,
		AppendCause: true,
	}
}

func NewFieldApiError(fieldError *FieldError) *ApiError {
	return &ApiError{
		Code:        InvalidFieldCode,
		Message:     InvalidFieldMessage,
		Status:      InvalidFieldStatus,
		Cause:       fieldError,
		AppendCause: true,
	}
}

func NewCouldNotValidate(err error) *ApiError {
	return &ApiError{
		Code:    CouldNotValidateCode,
		Message: CouldNotValidateMessage,
		Status:  CouldNotValidateStatus,
		Cause:   err,
	}
}
func NewUnauthorized() *ApiError {
	return &ApiError{
		Code:    UnauthorizedCode,
		Message: UnauthorizedMessage,
		Status:  UnauthorizedStatus,
	}
}

func NewInvalidFilter(cause error) *ApiError {
	return &ApiError{
		Code:        InvalidFilterCode,
		Message:     InvalidFilterMessage,
		Status:      InvalidFilterStatus,
		Cause:       cause,
		AppendCause: true,
	}
}

func NewInvalidPagination(err error) *ApiError {
	return &ApiError{
		Code:        InvalidPaginationCode,
		Message:     InvalidPaginationMessage,
		Status:      InvalidPaginationStatus,
		Cause:       err,
		AppendCause: true,
	}
}

func NewInvalidSort(err error) *ApiError {
	return &ApiError{
		Code:        InvalidSortCode,
		Message:     InvalidSortMessage,
		Status:      InvalidSortStatus,
		Cause:       err,
		AppendCause: true,
	}
}
