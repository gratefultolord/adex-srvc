package auction

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeStorage struct {
	partners []Partner
	err      error
}

func (s *fakeStorage) GetPartners(ctx context.Context) ([]Partner, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.partners, nil
}

type fakeDSPClient struct {
	send func(
		ctx context.Context,
		partner Partner,
		input Input,
	) error
}

func (c *fakeDSPClient) SendBidRequest(
	ctx context.Context,
	partner Partner,
	input Input,
) error {
	if c.send != nil {
		return c.send(ctx, partner, input)
	}

	return nil
}

func TestRunAuction(t *testing.T) {
	t.Run("storage error", func(t *testing.T) {
		expectedErr := errors.New("storage error")

		storage := &fakeStorage{
			err: expectedErr,
		}

		dspClient := &fakeDSPClient{}

		usecase := NewUsecase(
			storage,
			dspClient,
			200*time.Millisecond,
		)

		_, err := usecase.RunAuction(
			context.Background(),
			testInput(),
		)

		if err == nil {
			t.Fatal("RunAuction() error = nil, want error")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf(
				"RunAuction() error = %v, want wrapped %v",
				err,
				expectedErr,
			)
		}
	})

	t.Run("no matched partners", func(t *testing.T) {
		partner := testPartner()
		partner.IsEnabled = false

		storage := &fakeStorage{
			partners: []Partner{partner},
		}

		dspClient := &fakeDSPClient{}

		usecase := NewUsecase(
			storage,
			dspClient,
			200*time.Millisecond,
		)

		result, err := usecase.RunAuction(
			context.Background(),
			testInput(),
		)
		if err != nil {
			t.Fatalf("RunAuction() error = %v", err)
		}

		if result.RequestID != "request-1" {
			t.Errorf(
				"RequestID = %q, want %q",
				result.RequestID,
				"request-1",
			)
		}

		if result.Sent != 0 {
			t.Errorf("Sent = %d, want 0", result.Sent)
		}

		if result.Succeeded != 0 {
			t.Errorf(
				"Succeeded = %d, want 0",
				result.Succeeded,
			)
		}

		if len(result.MatchedDSPs) != 0 {
			t.Errorf(
				"len(MatchedDSPs) = %d, want 0",
				len(result.MatchedDSPs),
			)
		}
	})

	t.Run("all DSPs succeeded", func(t *testing.T) {
		partners := []Partner{
			testPartnerWithName("alpha", "DSP Alpha"),
			testPartnerWithName("beta", "DSP Beta"),
			testPartnerWithName("gamma", "DSP Gamma"),
		}

		storage := &fakeStorage{
			partners: partners,
		}

		dspClient := &fakeDSPClient{}

		usecase := NewUsecase(
			storage,
			dspClient,
			200*time.Millisecond,
		)

		result, err := usecase.RunAuction(
			context.Background(),
			testInput(),
		)
		if err != nil {
			t.Fatalf("RunAuction() error = %v", err)
		}

		if result.Sent != 3 {
			t.Errorf("Sent = %d, want 3", result.Sent)
		}

		if result.Succeeded != 3 {
			t.Errorf(
				"Succeeded = %d, want 3",
				result.Succeeded,
			)
		}

		if len(result.MatchedDSPs) != 3 {
			t.Fatalf(
				"len(MatchedDSPs) = %d, want 3",
				len(result.MatchedDSPs),
			)
		}

		wantNames := []string{
			"DSP Alpha",
			"DSP Beta",
			"DSP Gamma",
		}

		for i, want := range wantNames {
			if result.MatchedDSPs[i] != want {
				t.Errorf(
					"MatchedDSPs[%d] = %q, want %q",
					i,
					result.MatchedDSPs[i],
					want,
				)
			}
		}
	})

	t.Run("one DSP failed", func(t *testing.T) {
		partners := []Partner{
			testPartnerWithName("alpha", "DSP Alpha"),
			testPartnerWithName("beta", "DSP Beta"),
			testPartnerWithName("gamma", "DSP Gamma"),
		}

		storage := &fakeStorage{
			partners: partners,
		}

		dspClient := &fakeDSPClient{
			send: func(
				ctx context.Context,
				partner Partner,
				input Input,
			) error {
				if partner.UUID == "beta" {
					return errors.New("DSP failed")
				}

				return nil
			},
		}

		usecase := NewUsecase(
			storage,
			dspClient,
			200*time.Millisecond,
		)

		result, err := usecase.RunAuction(
			context.Background(),
			testInput(),
		)
		if err != nil {
			t.Fatalf("RunAuction() error = %v", err)
		}

		if result.Sent != 3 {
			t.Errorf("Sent = %d, want 3", result.Sent)
		}

		if result.Succeeded != 2 {
			t.Errorf(
				"Succeeded = %d, want 2",
				result.Succeeded,
			)
		}
	})

	t.Run("DSP timeout", func(t *testing.T) {
		const timeout = 30 * time.Millisecond

		storage := &fakeStorage{
			partners: []Partner{
				testPartnerWithName("alpha", "DSP Alpha"),
			},
		}

		dspClient := &fakeDSPClient{
			send: func(
				ctx context.Context,
				partner Partner,
				input Input,
			) error {
				<-ctx.Done()
				return ctx.Err()
			},
		}

		usecase := NewUsecase(
			storage,
			dspClient,
			timeout,
		)

		start := time.Now()

		result, err := usecase.RunAuction(
			context.Background(),
			testInput(),
		)

		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("RunAuction() error = %v", err)
		}

		if result.Sent != 1 {
			t.Errorf("Sent = %d, want 1", result.Sent)
		}

		if result.Succeeded != 0 {
			t.Errorf(
				"Succeeded = %d, want 0",
				result.Succeeded,
			)
		}

		if elapsed < timeout {
			t.Errorf(
				"RunAuction() returned too early: %v, timeout = %v",
				elapsed,
				timeout,
			)
		}

		if elapsed > 500*time.Millisecond {
			t.Errorf(
				"RunAuction() took too long: %v",
				elapsed,
			)
		}
	})
}

func TestRunAuctionCallsDSPsConcurrently(t *testing.T) {
	const partnersCount = 3

	partners := []Partner{
		testPartnerWithName("alpha", "DSP Alpha"),
		testPartnerWithName("beta", "DSP Beta"),
		testPartnerWithName("gamma", "DSP Gamma"),
	}

	storage := &fakeStorage{
		partners: partners,
	}

	var (
		mu      sync.Mutex
		started int
	)

	allStarted := make(chan struct{})
	var closeOnce sync.Once

	dspClient := &fakeDSPClient{
		send: func(
			ctx context.Context,
			partner Partner,
			input Input,
		) error {
			mu.Lock()
			started++

			if started == partnersCount {
				closeOnce.Do(func() {
					close(allStarted)
				})
			}

			mu.Unlock()

			select {
			case <-allStarted:
				return nil

			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}

	usecase := NewUsecase(
		storage,
		dspClient,
		200*time.Millisecond,
	)

	result, err := usecase.RunAuction(
		context.Background(),
		testInput(),
	)
	if err != nil {
		t.Fatalf("RunAuction() error = %v", err)
	}

	if result.Succeeded != partnersCount {
		t.Errorf(
			"Succeeded = %d, want %d",
			result.Succeeded,
			partnersCount,
		)
	}
}

func testInput() Input {
	return Input{
		RequestID:  "request-1",
		Country:    "RU",
		DeviceType: "mobile",
		BidFloor:   1.5,
		Categories: []string{"news"},
	}
}

func testPartner() Partner {
	return Partner{
		UUID:              "partner-1",
		Name:              "DSP Alpha",
		Endpoint:          "http://example.com/bid",
		IsEnabled:         true,
		Countries:         []string{"RU"},
		DeviceTypes:       []string{"mobile"},
		MinBidFloor:       1,
		BlockedCategories: []string{},
	}
}

func testPartnerWithName(uuid, name string) Partner {
	partner := testPartner()
	partner.UUID = uuid
	partner.Name = name

	return partner
}
