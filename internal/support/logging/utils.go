package logging

func checkMaxLimitCount(limit int) int {
	if limit > Configuration.Service.MaxResultCount || limit == 0 {
		return Configuration.Service.MaxResultCount
	}
	return limit
}
