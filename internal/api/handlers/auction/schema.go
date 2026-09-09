package auction

type AuctionRequest struct {
	RequestId  string   `json:"request_id"`
	Country    string   `json:"country"`
	DeviceType string   `json:"device_type"`
	BidFloor   float64  `json:"bid_floor"`
	Categories []string `json:"categories"`
}

type AuctionResponse struct {
	RequestId   string   `json:"request_id"`
	MatchedDSPs []string `json:"matched_dsps"`
	Sent        int      `json:"sent"`
	Succeeded   int      `json:"succeeded"`
	DurationMs  int      `json:"duration_ms"`
}
