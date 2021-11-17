package common

// CountResponse defines the Response Content for GET count DTO.
// This object and its properties correspond to the CountResponse object in the APIv2 specification:
// https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/core-data/2.1.0#/CountResponse
type CountResponse struct {
	BaseResponse `json:",inline"`
	Count        uint32
}

// NewCountResponse creates new CountResponse with all fields set appropriately
func NewCountResponse(requestId string, message string, statusCode int, count uint32) CountResponse {
	return CountResponse{
		BaseResponse: NewBaseResponse(requestId, message, statusCode),
		Count:        count,
	}
}
