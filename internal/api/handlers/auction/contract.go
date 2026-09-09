package auction

import (
	"context"

	"github.com/gratefultolord/adex-srvc/internal/usecases/auction"
)

type usecase interface {
	RunAuction(cxt context.Context, req AuctionRequest) (auction.AuctionDTO, error)
}
