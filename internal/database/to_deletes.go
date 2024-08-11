package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ToDelete struct {
	ID         uuid.UUID `db:"id"`
	FileID     string    `db:"image_id"`
	DeleteTime time.Time `db:"delete_time"`
	Created    time.Time `db:"created"`
	Updated    time.Time `db:"updated"`
}

func (db *DB) InsertToDelete(pdfId string, deleteTime time.Time) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var id uuid.UUID

	query := `
		INSERT INTO to_deletes (pdf_id, delete_time)
		VALUES ($1, $2)
		returning id`

	err := db.GetContext(ctx, &id, query, pdfId, deleteTime)
	if err != nil {
		return uuid.Nil, err
	}

	return id, err
}
