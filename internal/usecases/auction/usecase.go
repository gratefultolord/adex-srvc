package auction

import (
	"context"
	"fmt"
	"time"
)

type Usecase struct {
	storage    storage
	dspClient  dspClient
	dspTimeout time.Duration
}

func NewUsecase(storage storage, dspClient dspClient, dspTimeout time.Duration) *Usecase {
	return &Usecase{
		storage:    storage,
		dspClient:  dspClient,
		dspTimeout: dspTimeout,
	}
}

func (u *Usecase) RunAuction(ctx context.Context, req Input) (Result, error) {
	start := time.Now()

	partners, err := u.storage.GetPartners(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("u.storage.GetPartners: %w", err)
	}

	filteredPartners := FilterPartners(req, partners)

	result := Result{
		RequestID:   req.RequestID,
		MatchedDSPs: make([]string, 0, len(filteredPartners)),
		Sent:        len(filteredPartners),
	}

	for _, partner := range filteredPartners {
		result.MatchedDSPs = append(
			result.MatchedDSPs,
			partner.Name,
		)
	}

	if len(filteredPartners) == 0 {
		result.DurationMs = int(time.Since(start).Milliseconds())
		return result, nil
	}

	auctionCtx, cancel := context.WithTimeout(ctx, u.dspTimeout)
	defer cancel()

	resultsCh := make(chan dspResult, len(filteredPartners))

	for _, partner := range filteredPartners {
		go func(p Partner) {
			err := u.dspClient.SendBidRequest(auctionCtx, p, req)

			resultsCh <- dspResult{
				partner: p,
				err:     err,
			}
		}(partner)
	}

	for range filteredPartners {
		select {
		case res := <-resultsCh:
			if res.err == nil {
				result.Succeeded++
			}

		case <-auctionCtx.Done():
			result.DurationMs = int(time.Since(start).Milliseconds())
			return result, nil
		}
	}

	result.DurationMs = int(time.Since(start).Milliseconds())
	return result, nil
}
