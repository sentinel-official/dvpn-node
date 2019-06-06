package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // SQLite driver
)

type DB struct {
	db *gorm.DB
}

func NewDatabase(source string) (*DB, error) {
	db, err := gorm.Open("sqlite3", source)
	if err != nil {
		return nil, err
	}

	db.LogMode(false)

	db.AutoMigrate(&subscription{})
	db.AutoMigrate(&session{})

	db.Exec("PRAGMA FOREIGN_KEYS = ON")
	db.Exec("PRAGMA JOURNAL_MODE = WAL")
	db.Exec("PRAGMA SYNCHRONOUS = NORMAL")

	return &DB{
		db: db,
	}, nil
}
