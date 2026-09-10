package storage

import (
	"context"
	"fmt"

	"github.com/gratefultolord/adex-srvc/internal/usecases/auction"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
)

type Storage struct {
	db *sqlx.DB
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) GetPartners(ctx context.Context) ([]auction.Partner, error) {
	var partners []PartnerDAO

	query := `
        SELECT
            uuid,
            name,
            endpoint,
            is_enabled,
            countries,
            device_types,
            min_bid_floor,
            blocked_categories
        FROM partners
		ORDER BY name
    `

	if err := s.db.SelectContext(ctx, &partners, query); err != nil {
		return nil, fmt.Errorf("s.db.SelectContext: %w", err)
	}

	return lo.Map(partners, func(partner PartnerDAO, _ int) auction.Partner {
		return auction.Partner{
			UUID:              partner.UUID,
			Name:              partner.Name,
			Endpoint:          partner.Endpoint,
			IsEnabled:         partner.IsEnabled,
			Countries:         []string(partner.Countries),
			DeviceTypes:       []string(partner.DeviceTypes),
			MinBidFloor:       partner.MinBidFloor,
			BlockedCategories: []string(partner.BlockedCategories),
		}
	}), nil
}
