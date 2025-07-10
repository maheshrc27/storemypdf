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

	uid := uuid.New()

	query := `
		INSERT INTO to_deletes (id, file_id, delete_time, created, updated)
		VALUES ($1, $2, $3, $4, $5)
		returning id`

	err := db.GetContext(ctx, &id, query, uid.String(), fileId, deleteTime, time.Now(), time.Now())
	if err != nil {
		return uuid.Nil, err
	}

	return id, err
}

func (db *DB) GetToDeletes() ([]ToDelete, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var todeletes []ToDelete

	query := `SELECT * FROM to_deletes WHERE delete_time < CURRENT_TIMESTAMP`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var td ToDelete
		if err := rows.Scan(&td.ID, &td.FileID, &td.DeleteTime, &td.Created, &td.Updated); err != nil {
			return todeletes, false, err
		}
		todeletes = append(todeletes, td)
	}

	if err = rows.Err(); err != nil {
		return nil, false, err
	}

	if len(todeletes) == 0 {
		return nil, false, nil
	}

	return todeletes, true, nil
}
