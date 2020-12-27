package apiv1

type ReqStatResponse struct {
	IP               string `json:"ip"`
	RequestCount     int64  `json:"requestCount"`
	RemainingRequest int64  `json:"remainingRequest"`
	ResetAfter       string `json:"resetAfter"`
	ResetAt          int64  `json:"resetAt"`
}

type GetGuardianAnimalRequest struct {
	Name string `query:"name"`
}

type GetGuardianAnimalResponse struct {
	Animal      string `json:"animal"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
