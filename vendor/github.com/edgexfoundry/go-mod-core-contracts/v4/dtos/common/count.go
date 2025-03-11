package common

// CountResponse defines the Response Content for GET count DTO.
type CountResponse struct {
	BaseResponse `json:",inline"`
	Count        uint32 `json:"count"`
}

// NewCountResponse creates new CountResponse with all fields set appropriately
func NewCountResponse(requestId string, message string, statusCode int, count uint32) CountResponse {
	return CountResponse{
		BaseResponse: NewBaseResponse(requestId, message, statusCode),
		Count:        count,
	}
}
