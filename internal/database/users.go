package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `db:"id"`
	Email          string    `db:"email"`
	HashedPassword string    `db:"hashed_password"`
	Verified       bool      `db:"verified"`
	Created        time.Time `db:"created"`
	Updated        time.Time `db:"updated"`
}

func (db *DB) InsertUser(email, hashedPassword string) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var id uuid.UUID

	query := `
		INSERT INTO users (created, email, hashed_password, verified)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	err := db.GetContext(ctx, &id, query, time.Now(), email, hashedPassword, false)
	if err != nil {
		return uuid.Nil, err
	}

	return id, err
}

func (db *DB) GetUser(id uuid.UUID) (*User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var user User

	query := `SELECT * FROM users WHERE id = $1`

	err := db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &user, true, err
}

func (db *DB) GetUserByEmail(email string) (*User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var user User

	query := `SELECT * FROM users WHERE email = $1`

	err := db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &user, true, nil
}

func (db *DB) UpdateUserHashedPassword(id uuid.UUID, hashedPassword string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE users SET hashed_password = $1, updated = $2 WHERE id = $3`

	_, err := db.ExecContext(ctx, query, hashedPassword, time.Now(), id)
	return err
}

func (db *DB) UpdateVerification(id uuid.UUID, verified bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE users SET verified = $1, updated = $2 WHERE id = $3`

	_, err := db.ExecContext(ctx, query, verified, time.Now(), id)
	return err
}

func (db *DB) DeleteUser(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `DELETE FROM users WHERE id = $1`

	_, err := db.ExecContext(ctx, query, id)
	return err
}
