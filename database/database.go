package database

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type DB struct {
	db *gorm.DB
}

func NewDatabase(source string) (*DB, error) {
	db, err := gorm.Open("sqlite3", source)
	if err != nil {
		return nil, err
	}

	return &DB{
		db: db,
	}, nil
}
