package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID          string    `db:"id"`
	FileName    string    `db:"filename"`
	Description string    `db:"description"`
	FileType    string    `db:"file_type"`
	Size        int64     `db:"size"`
	UserID      uuid.UUID `db:"user_id"`
	Created     time.Time `db:"created"`
	Updated     time.Time `db:"updated"`
}

func (db *DB) InsertFile(id, filename, description, fileType string, size int64, userId uuid.UUID) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO files (id, filename, description, file_type, size, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.ExecContext(ctx, query, id, filename, description, fileType, size, userId)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (db *DB) GetFile(id string) (*File, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var image File

	query := `SELECT * FROM files WHERE id = $1`

	err := db.GetContext(ctx, &image, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &image, true, err
}

func (db *DB) GetFilesByUserID(userId uuid.UUID) ([]File, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var files []File

	query := `SELECT * FROM files WHERE user_id = $1`

	rows, err := db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	for rows.Next() {
		var file File
		if err := rows.Scan(&file.ID, &file.FileName, &file.Description, &file.FileType, &file.Size,
			&file.UserID, &file.Created, &file.Updated); err != nil {
			return files, false, err
		}
		files = append(files, file)
	}

	if err = rows.Err(); err != nil {
		return nil, false, err
	}

	if len(files) == 0 {
		return nil, false, nil
	}

	return files, true, nil
}

func (db *DB) DeleteFile(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `DELETE FROM files WHERE id = $1`

	_, err := db.ExecContext(ctx, query, id)
	return err
}
