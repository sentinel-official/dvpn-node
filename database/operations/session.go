package operations

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/sentinel-official/sentinel-dvpnx/database/models"
)

// SessionInsertOne inserts a single Session record into the database.
func SessionInsertOne(db *gorm.DB, session *models.Session) error {
	fn := func(db *gorm.DB) error {
		if err := db.Create(session).Error; err != nil {
			return fmt.Errorf("inserting session: %w", err)
		}

		return nil
	}

	if err := db.Transaction(fn); err != nil {
		return fmt.Errorf("running tx: %w", err)
	}

	return nil
}

// SessionInsertMany inserts multiple Session records into the database.
func SessionInsertMany(db *gorm.DB, sessions []models.Session) error {
	fn := func(db *gorm.DB) error {
		if err := db.Create(sessions).Error; err != nil {
			return fmt.Errorf("inserting sessions: %w", err)
		}

		return nil
	}

	if err := db.Transaction(fn); err != nil {
		return fmt.Errorf("running tx: %w", err)
	}

	return nil
}

// SessionFindOne retrieves a single session record from the database based on the provided query.
func SessionFindOne(db *gorm.DB, query map[string]any) (session *models.Session, err error) {
	db = applyQuery(db, query)
	if err := db.First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("finding session with query %v: %w", query, err)
	}

	return session, nil
}

// SessionFind retrieves multiple session records from the database based on the provided query.
func SessionFind(db *gorm.DB, query map[string]any) (sessions []models.Session, err error) {
	db = applyQuery(db, query)
	if err := db.Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("finding sessions with query %v: %w", query, err)
	}

	return sessions, nil
}

// SessionFindOneAndUpdate finds a single session record based on the provided query and updates it with the provided updates.
func SessionFindOneAndUpdate(db *gorm.DB, query, updates map[string]any) (session *models.Session, err error) {
	fn := func(db *gorm.DB) error {
		session, err = SessionFindOne(db, query)
		if err != nil {
			return fmt.Errorf("finding session with query %v for update: %w", query, err)
		}

		if session == nil {
			return nil
		}

		if err := db.Model(session).Updates(updates).Error; err != nil {
			return fmt.Errorf("updating session with query %v: %w", query, err)
		}

		return nil
	}

	if err := db.Transaction(fn); err != nil {
		return nil, fmt.Errorf("running tx: %w", err)
	}

	return session, nil
}

// SessionUpdateMany updates multiple session records based on the provided query and updates them with the provided updates.
func SessionUpdateMany(db *gorm.DB, query map[string]any, updates map[string]any) error {
	fn := func(db *gorm.DB) error {
		db = applyQuery(db, query)
		if err := db.Model(&models.Session{}).Updates(updates).Error; err != nil {
			return fmt.Errorf("updating sessions with query %v: %w", query, err)
		}

		return nil
	}

	if err := db.Transaction(fn); err != nil {
		return fmt.Errorf("running tx: %w", err)
	}

	return nil
}

// SessionFindOneAndDelete finds a single session record based on the provided query and deletes it.
func SessionFindOneAndDelete(db *gorm.DB, query map[string]any) (session *models.Session, err error) {
	fn := func(db *gorm.DB) error {
		session, err = SessionFindOne(db, query)
		if err != nil {
			return fmt.Errorf("finding session with query %v for deletion: %w", query, err)
		}

		if session == nil {
			return nil
		}

		if err := db.Model(session).Delete(nil).Error; err != nil {
			return fmt.Errorf("deleting session with query %v: %w", query, err)
		}

		return nil
	}

	if err := db.Transaction(fn); err != nil {
		return nil, fmt.Errorf("running tx: %w", err)
	}

	return session, nil
}

// SessionCount returns the number of active sessions stored.
func SessionCount(db *gorm.DB) (int, error) {
	var count int64
	if err := db.Model(&models.Session{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("counting sessions: %w", err)
	}

	return int(count), nil
}

// SessionDeleteMany deletes multiple session records based on the provided query.
func SessionDeleteMany(db *gorm.DB, query map[string]any) error {
	fn := func(db *gorm.DB) error {
		db = applyQuery(db, query)
		if err := db.Model(&models.Session{}).Delete(nil).Error; err != nil {
			return fmt.Errorf("deleting sessions with query %v: %w", query, err)
		}

		return nil
	}

	if err := db.Transaction(fn); err != nil {
		return fmt.Errorf("running tx: %w", err)
	}

	return nil
}
