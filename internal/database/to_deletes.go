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

func (db *DB) InsertToDelete(fileId string, deleteTime time.Time) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var id uuid.UUID

	query := `
		INSERT INTO to_deletes (file_id, delete_time)
		VALUES ($1, $2)
		returning id`

	err := db.GetContext(ctx, &id, query, fileId, deleteTime)
	if err != nil {
		return uuid.Nil, err
	}

	return id, err
}
