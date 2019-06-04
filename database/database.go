package database

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

	db.AutoMigrate(&subscription{})
	db.AutoMigrate(&session{})

	db.Model(&session{}).AddForeignKey("_id", "subscriptions(_id)", "CASCADE", "CASCADE")

	return &DB{
		db: db,
	}, nil
}
