package auction

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	auctionUC "github.com/gratefultolord/adex-srvc/internal/usecases/auction"
	"go.uber.org/zap"
)

type Handler struct {
	logger  *zap.Logger
	usecase usecase
}

func NewHandler(logger *zap.Logger, usecase usecase) *Handler {
	return &Handler{
		logger:  logger,
		usecase: usecase,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	var req AuctionRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
		return
	}

	if err := validateAuctionRequest(req); err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
		return
	}

	result, err := h.usecase.RunAuction(r.Context(), auctionUC.Input{
		RequestID:  req.RequestID,
		Country:    req.Country,
		DeviceType: req.DeviceType,
		BidFloor:   req.BidFloor,
		Categories: req.Categories,
	})
	if err != nil {
		h.logger.Error(
			"h.usecase.RunAuction",
			zap.Error(err),
		)

		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	resp := AuctionResponse{
		RequestID:   result.RequestID,
		MatchedDSPs: result.MatchedDSPs,
		Sent:        result.Sent,
		Succeeded:   result.Succeeded,
		DurationMs:  result.DurationMs,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error(
			"json.NewEncoder.Encode",
			zap.Error(err),
		)
	}
}

func validateAuctionRequest(req AuctionRequest) error {
	if strings.TrimSpace(req.RequestID) == "" {
		return errors.New("request_id is required")
	}

	if strings.TrimSpace(req.Country) == "" {
		return errors.New("country is required")
	}

	if strings.TrimSpace(req.DeviceType) == "" {
		return errors.New("device_type is required")
	}

	if req.BidFloor < 0 {
		return errors.New("bid_floor must be non-negative")
	}

	return nil
}
