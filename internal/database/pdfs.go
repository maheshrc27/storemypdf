package database

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Pdf struct {
	ID          string    `db:"id"`
	FileName    string    `db:"filename"`
	Description string    `db:"description"`
	UserID      int       `db:"user_id"`
	Created     time.Time `db:"created"`
	Updated     time.Time `db:"updated"`
}

func (db *DB) InsertPdf(id, filename, description string, width, height, userId int) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO pdfs (id, filename, description, user_id, updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := db.ExecContext(ctx, query, id, filename, description, width, height, userId, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetPdf(id string) (*Pdf, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var image Pdf

	query := `SELECT * FROM pdfs WHERE id = $1`

	err := db.GetContext(ctx, &image, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &image, true, err
}

func (db *DB) GetPdfsByUserID(userId string) ([]Pdf, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var images []Pdf

	query := `SELECT * FROM pdfs WHERE user_id = $1`

	rows, err := db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var img Pdf
		if err := rows.Scan(&img.ID, &img.FileName, &img.Description,
			&img.UserID, &img.Created, &img.Updated); err != nil {
			return images, false, err
		}
		images = append(images, img)
	}

	if err = rows.Err(); err != nil {
		return nil, false, err
	}

	if len(images) == 0 {
		return nil, false, nil
	}

	return images, true, nil
}

func (db *DB) DeletePdf(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `DELETE FROM pdfs WHERE id = $1`

	_, err := db.ExecContext(ctx, query, id)
	return err
}
