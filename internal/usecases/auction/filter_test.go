package auction

import "testing"

func TestMatchPartner(t *testing.T) {
	tests := []struct {
		name    string
		input   Input
		partner Partner
		want    bool
	}{
		{
			name:    "partner matches all conditions",
			input:   defaultInput(),
			partner: defaultPartner(),
			want:    true,
		},
		{
			name:  "partner is disabled",
			input: defaultInput(),
			partner: func() Partner {
				p := defaultPartner()
				p.IsEnabled = false
				return p
			}(),
			want: false,
		},
		{
			name:  "country is allowed",
			input: defaultInput(),
			partner: func() Partner {
				p := defaultPartner()
				p.Countries = []string{"KZ", "RU"}
				return p
			}(),
			want: true,
		},
		{
			name:  "country is not allowed",
			input: defaultInput(),
			partner: func() Partner {
				p := defaultPartner()
				p.Countries = []string{"US", "KZ"}
				return p
			}(),
			want: false,
		},
		{
			name:  "empty countries allows any country",
			input: defaultInput(),
			partner: func() Partner {
				p := defaultPartner()
				p.Countries = []string{}
				return p
			}(),
			want: true,
		},
		{
			name:  "device type is allowed",
			input: defaultInput(),
			partner: func() Partner {
				p := defaultPartner()
				p.DeviceTypes = []string{"desktop", "mobile"}
				return p
			}(),
			want: true,
		},
		{
			name:  "device type is not allowed",
			input: defaultInput(),
			partner: func() Partner {
				p := defaultPartner()
				p.DeviceTypes = []string{"desktop"}
				return p
			}(),
			want: false,
		},
		{
			name:  "empty device types allows any device",
			input: defaultInput(),
			partner: func() Partner {
				p := defaultPartner()
				p.DeviceTypes = []string{}
				return p
			}(),
			want: true,
		},
		{
			name:  "bid floor is lower than minimum",
			input: defaultInput(),
			partner: func() Partner {
				p := defaultPartner()
				p.MinBidFloor = 2.0
				return p
			}(),
			want: false,
		},
		{
			name: "bid floor equals minimum",
			input: func() Input {
				in := defaultInput()
				in.BidFloor = 1.5
				return in
			}(),
			partner: func() Partner {
				p := defaultPartner()
				p.MinBidFloor = 1.5
				return p
			}(),
			want: true,
		},
		{
			name: "bid floor is greater than minimum",
			input: func() Input {
				in := defaultInput()
				in.BidFloor = 2.0
				return in
			}(),
			partner: func() Partner {
				p := defaultPartner()
				p.MinBidFloor = 1.0
				return p
			}(),
			want: true,
		},
		{
			name: "request contains blocked category",
			input: func() Input {
				in := defaultInput()
				in.Categories = []string{"news", "gambling"}
				return in
			}(),
			partner: func() Partner {
				p := defaultPartner()
				p.BlockedCategories = []string{"gambling"}
				return p
			}(),
			want: false,
		},
		{
			name: "request does not contain blocked category",
			input: func() Input {
				in := defaultInput()
				in.Categories = []string{"news", "sport"}
				return in
			}(),
			partner: func() Partner {
				p := defaultPartner()
				p.BlockedCategories = []string{"gambling", "adult"}
				return p
			}(),
			want: true,
		},
		{
			name: "empty request categories",
			input: func() Input {
				in := defaultInput()
				in.Categories = []string{}
				return in
			}(),
			partner: defaultPartner(),
			want:    true,
		},
		{
			name:  "empty blocked categories",
			input: defaultInput(),
			partner: func() Partner {
				p := defaultPartner()
				p.BlockedCategories = []string{}
				return p
			}(),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchPartner(tt.input, tt.partner)

			if got != tt.want {
				t.Errorf(
					"MatchPartner() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestFilterPartners(t *testing.T) {
	input := defaultInput()

	partners := []Partner{
		{
			UUID:              "alpha",
			Name:              "DSP Alpha",
			IsEnabled:         true,
			Countries:         []string{"RU"},
			DeviceTypes:       []string{"mobile"},
			MinBidFloor:       0.5,
			BlockedCategories: []string{},
		},
		{
			UUID:              "beta",
			Name:              "DSP Beta",
			IsEnabled:         false,
			Countries:         []string{"RU"},
			DeviceTypes:       []string{"mobile"},
			MinBidFloor:       0.5,
			BlockedCategories: []string{},
		},
		{
			UUID:              "gamma",
			Name:              "DSP Gamma",
			IsEnabled:         true,
			Countries:         []string{},
			DeviceTypes:       []string{},
			MinBidFloor:       1.0,
			BlockedCategories: []string{"gambling"},
		},
	}

	got := FilterPartners(input, partners)

	if len(got) != 2 {
		t.Fatalf(
			"FilterPartners() returned %d partners, want 2",
			len(got),
		)
	}

	if got[0].UUID != "alpha" {
		t.Errorf(
			"first partner UUID = %q, want %q",
			got[0].UUID,
			"alpha",
		)
	}

	if got[1].UUID != "gamma" {
		t.Errorf(
			"second partner UUID = %q, want %q",
			got[1].UUID,
			"gamma",
		)
	}
}

func defaultInput() Input {
	return Input{
		RequestID:  "request-1",
		Country:    "RU",
		DeviceType: "mobile",
		BidFloor:   1.5,
		Categories: []string{"news", "sport"},
	}
}

func defaultPartner() Partner {
	return Partner{
		UUID:              "partner-1",
		Name:              "DSP Alpha",
		Endpoint:          "http://localhost:9001/bid",
		IsEnabled:         true,
		Countries:         []string{"RU"},
		DeviceTypes:       []string{"mobile"},
		MinBidFloor:       1.0,
		BlockedCategories: []string{"gambling"},
	}
}
