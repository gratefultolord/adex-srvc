package auction

import (
	"context"

	auctionUC "github.com/gratefultolord/adex-srvc/internal/usecases/auction"
)

type usecase interface {
	RunAuction(ctx context.Context, req auctionUC.Input) (auctionUC.Result, error)
}
