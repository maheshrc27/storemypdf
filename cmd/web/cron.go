package main

import (
	"log"
	"sync"

	"github.com/maheshrc27/storemypdf/internal/database"
)

// DeleteFileScheduler periodically deletes files marked for deletion.
func (app *application) DeleteFileScheduler() {

	todeletes, found, err := app.db.GetToDeletes()
	if err != nil {
		log.Printf("failed to get files to delete: %v", err)
		return
	}

	if !found {
		return
	}

	var wg sync.WaitGroup

	for _, v := range todeletes {
		fileid := v.FileID
		wg.Add(1)
		go func(fileid string) {
			defer wg.Done()
			if err := DeleteUploads(app.db, fileid); err != nil {
				log.Printf("failed to delete file %s: %v", fileid, err)
			}
		}(fileid)
	}

	wg.Wait()
	log.Println("all file deletion tasks completed")
}

func DeleteUploads(db *database.DB, fileid string) error {
	err := DeleteS3Object(fileid)
	if err != nil {
		return err
	}

	if err := db.DeleteFile(fileid); err != nil {
		return err
	}

	return nil
}
