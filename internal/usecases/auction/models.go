package auction

type Input struct {
	RequestID  string
	Country    string
	DeviceType string
	BidFloor   float64
	Categories []string
}

type Result struct {
	RequestID   string
	MatchedDSPs []string
	Sent        int
	Succeeded   int
	DurationMs  int
}

type Partner struct {
	UUID              string
	Name              string
	Endpoint          string
	IsEnabled         bool
	Countries         []string
	DeviceTypes       []string
	MinBidFloor       float64
	BlockedCategories []string
}

type dspResult struct {
	partner Partner
	err     error
}
