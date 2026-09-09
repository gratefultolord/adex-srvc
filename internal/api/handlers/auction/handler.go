package auction

import (
	"encoding/json"
	"net/http"
	"strings"

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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.RequestId) == "" ||
		strings.TrimSpace(req.Country) == "" ||
		strings.TrimSpace(req.DeviceType) == "" ||
		req.BidFloor < 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	result, err := h.usecase.RunAuction(r.Context(), req)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("run auction", zap.Error(err))
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := AuctionResponse{
		RequestId:   result.RequestId,
		MatchedDSPs: result.MatchedDSPs,
		Sent:        result.Sent,
		Succeeded:   result.Succeeded,
		DurationMs:  result.DurationMs,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		if h.logger != nil {
			h.logger.Error("encode auction response", zap.Error(err))
		}
	}
}
