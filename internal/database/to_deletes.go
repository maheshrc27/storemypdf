package database

import (
	"context"
	"time"
)

type ToDelete struct {
	ID         string    `db:"id"`
	ImageID    string    `db:"image_id"`
	DeleteTime time.Time `db:"delete_time"`
	Created    time.Time `db:"created"`
	Updated    time.Time `db:"updated"`
}

func (db *DB) InsertToDelete(pdfId string, deleteTime time.Time) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO to_deletes (pdf_id, delete_time)
		VALUES ($1, $2)`

	result, err := db.ExecContext(ctx, query, pdfId, deleteTime)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), err
}
