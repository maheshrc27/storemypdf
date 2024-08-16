package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Key struct {
	ID         int8      `db:"id"`
	ApiKeyHash string    `db:"api_key_hash"`
	UserID     uuid.UUID `db:"user_id"`
	Active     bool      `db:"active"`
	Created    time.Time `db:"created"`
	Updated    time.Time `db:"updated"`
}

func (db *DB) InsertKey(keyHash string, userID uuid.UUID) (int8, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var id int8

	query := `
		INSERT INTO api_keys (api_key_hash, user_id)
		VALUES ($1, $2)
		RETURNING id`

	err := db.GetContext(ctx, &id, query, keyHash, userID)
	if err != nil {
		return 0, err
	}

	return id, err
}

func (db *DB) GetKey(id int8) (*Key, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var key Key

	query := `SELECT * FROM api_keys WHERE id = $1`

	err := db.GetContext(ctx, &key, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &key, true, err
}

func (db *DB) GetUserIDByKey(key string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var userid string

	query := `SELECT user_id FROM api_keys WHERE api_key_hash = $1`

	err := db.GetContext(ctx, &userid, query, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}

	return userid, true, err
}

func (db *DB) GetKeysByUserID(userId uuid.UUID) ([]Key, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var keys []Key

	query := `SELECT * FROM api_keys WHERE user_id = $1`

	rows, err := db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var key Key
		if err := rows.Scan(&key.ID, &key.ApiKeyHash, &key.UserID, &key.Active,
			&key.Created, &key.Updated); err != nil {
			return keys, false, err
		}
		keys = append(keys, key)
	}

	if err = rows.Err(); err != nil {
		return nil, false, err
	}

	if len(keys) == 0 {
		return nil, false, nil
	}

	return keys, true, nil
}

func (db *DB) UpdateKeyHash(id int8, keyHash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE api_keys SET api_key_hash = $1, updated = $2 WHERE id = $3`

	_, err := db.ExecContext(ctx, query, keyHash, time.Now(), id)
	return err
}

func (db *DB) UpdateKeyStatus(id int8, active bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE api_keys SET active = $1, updated = $2 WHERE id = $3`

	_, err := db.ExecContext(ctx, query, active, time.Now(), id)
	return err
}

func (db *DB) DeleteKey(id int8) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `DELETE FROM api_keys WHERE id = $1`

	_, err := db.ExecContext(ctx, query, id)
	return err
}
