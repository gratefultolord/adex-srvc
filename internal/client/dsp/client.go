package dsp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gratefultolord/adex-srvc/internal/usecases/auction"
)

type Client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
	}
}

type bidRequest struct {
	RequestID  string   `json:"request_id"`
	Country    string   `json:"country"`
	DeviceType string   `json:"device_type"`
	BidFloor   float64  `json:"bid_floor"`
	Categories []string `json:"categories"`
}

func (c *Client) SendBidRequest(
	ctx context.Context,
	partner auction.Partner,
	input auction.Input,
) error {
	body, err := json.Marshal(bidRequest{
		RequestID:  input.RequestID,
		Country:    input.Country,
		DeviceType: input.DeviceType,
		BidFloor:   input.BidFloor,
		Categories: input.Categories,
	})
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		partner.Endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("c.httpClient.Do: %w", err)
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("DSP returned status %d", resp.StatusCode)
	}
	return nil
}
