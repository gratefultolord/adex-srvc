package auction

type AuctionDTO struct {
	RequestId   string
	MatchedDSPs []string
	Sent        int
	Succeeded   int
	DurationMs  int
}
