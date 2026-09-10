package dsp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	auctionUC "github.com/gratefultolord/adex-srvc/internal/usecases/auction"
)

func TestClient_SendBidRequest(t *testing.T) {
	t.Run("sends correct request", func(t *testing.T) {
		input := testInput()

		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf(
						"method = %q, want %q",
						r.Method,
						http.MethodPost,
					)
				}

				if got := r.Header.Get("Content-Type"); got != "application/json" {
					t.Errorf(
						"Content-Type = %q, want %q",
						got,
						"application/json",
					)
				}

				var got bidRequest

				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Fatalf("json.NewDecoder.Decode: %v", err)
				}

				if got.RequestID != input.RequestID {
					t.Errorf(
						"RequestID = %q, want %q",
						got.RequestID,
						input.RequestID,
					)
				}

				if got.Country != input.Country {
					t.Errorf(
						"Country = %q, want %q",
						got.Country,
						input.Country,
					)
				}

				if got.DeviceType != input.DeviceType {
					t.Errorf(
						"DeviceType = %q, want %q",
						got.DeviceType,
						input.DeviceType,
					)
				}

				if got.BidFloor != input.BidFloor {
					t.Errorf(
						"BidFloor = %v, want %v",
						got.BidFloor,
						input.BidFloor,
					)
				}

				if len(got.Categories) != len(input.Categories) {
					t.Fatalf(
						"len(Categories) = %d, want %d",
						len(got.Categories),
						len(input.Categories),
					)
				}

				for i := range input.Categories {
					if got.Categories[i] != input.Categories[i] {
						t.Errorf(
							"Categories[%d] = %q, want %q",
							i,
							got.Categories[i],
							input.Categories[i],
						)
					}
				}

				w.WriteHeader(http.StatusOK)
			}),
		)
		defer server.Close()

		client := NewClient(server.Client())

		partner := testPartner(server.URL)

		err := client.SendBidRequest(
			context.Background(),
			partner,
			input,
		)
		if err != nil {
			t.Fatalf("SendBidRequest() error = %v", err)
		}
	})

	t.Run("accepts any 2xx response", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}),
		)
		defer server.Close()

		client := NewClient(server.Client())

		err := client.SendBidRequest(
			context.Background(),
			testPartner(server.URL),
			testInput(),
		)
		if err != nil {
			t.Fatalf("SendBidRequest() error = %v", err)
		}
	})

	t.Run("returns error on 5xx response", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
			}),
		)
		defer server.Close()

		client := NewClient(server.Client())

		err := client.SendBidRequest(
			context.Background(),
			testPartner(server.URL),
			testInput(),
		)

		if err == nil {
			t.Fatal("SendBidRequest() error = nil, want error")
		}
	})

	t.Run("respects context deadline", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				select {
				case <-r.Context().Done():
					return

				case <-time.After(time.Second):
					w.WriteHeader(http.StatusOK)
				}
			}),
		)
		defer server.Close()

		client := NewClient(server.Client())

		ctx, cancel := context.WithTimeout(
			context.Background(),
			30*time.Millisecond,
		)
		defer cancel()

		start := time.Now()

		err := client.SendBidRequest(
			ctx,
			testPartner(server.URL),
			testInput(),
		)

		elapsed := time.Since(start)

		if err == nil {
			t.Fatal("SendBidRequest() error = nil, want timeout error")
		}

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf(
				"SendBidRequest() error = %v, want context.DeadlineExceeded",
				err,
			)
		}

		if elapsed > 500*time.Millisecond {
			t.Errorf(
				"SendBidRequest() ignored context deadline: took %v",
				elapsed,
			)
		}
	})
}

func testInput() auctionUC.Input {
	return auctionUC.Input{
		RequestID:  "request-1",
		Country:    "RU",
		DeviceType: "mobile",
		BidFloor:   1.5,
		Categories: []string{"news", "sport"},
	}
}

func testPartner(endpoint string) auctionUC.Partner {
	return auctionUC.Partner{
		UUID:              "partner-1",
		Name:              "DSP Alpha",
		Endpoint:          endpoint,
		IsEnabled:         true,
		Countries:         []string{"RU"},
		DeviceTypes:       []string{"mobile"},
		MinBidFloor:       1,
		BlockedCategories: []string{},
	}
}
