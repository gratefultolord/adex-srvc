package auction

import (
	"context"
)

type storage interface {
	GetPartners(ctx context.Context) ([]Partner, error)
}

type dspClient interface {
	SendBidRequest(ctx context.Context, partner Partner, input Input) error
}
