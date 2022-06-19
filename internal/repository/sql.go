package repository

import (
	"database/sql"
	"project/internal/consts"
)

func Error(err error) error {
	switch err {
	case sql.ErrNoRows:
		return consts.ErrNotFound
	default:
		return err
	}
}
