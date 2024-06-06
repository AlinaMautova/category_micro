package storage

import (
	"catalogue-service/internal/lib"
	"database/sql"
	"fmt"
)

func New(dsn string) (*lib.ItemRepo, error) {
	const op = "data.sqlite.New"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &lib.ItemRepo{DB: db}, nil
}
