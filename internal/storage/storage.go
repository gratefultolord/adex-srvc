package storage

import "github.com/jmoiron/sqlx"

type Storage struct {
	db *sqlx.DB
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) ListPartners() ([]string, error) {
	partners := make([]string, 0)

	return partners, nil
}
