package storage

import (
	"github.com/lib/pq"
)

type PartnerDAO struct {
	UUID              string         `db:"uuid"`
	Name              string         `db:"name"`
	Endpoint          string         `db:"endpoint"`
	IsEnabled         bool           `db:"is_enabled"`
	Countries         pq.StringArray `db:"countries"`
	DeviceTypes       pq.StringArray `db:"device_types"`
	MinBidFloor       float64        `db:"min_bid_floor"`
	BlockedCategories pq.StringArray `db:"blocked_categories"`
}
