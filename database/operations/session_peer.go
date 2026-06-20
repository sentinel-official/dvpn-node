package operations

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/sentinel-official/sentinel-dvpnx/database/models"
)

// SessionPeerInsertOne inserts a single SessionPeer record into the database.
func SessionPeerInsertOne(db *gorm.DB, peer *models.SessionPeer) error {
	fn := func(db *gorm.DB) error {
		if err := db.Create(peer).Error; err != nil {
			return fmt.Errorf("inserting session_peer: %w", err)
		}

		return nil
	}

	if err := db.Transaction(fn); err != nil {
		return fmt.Errorf("running tx: %w", err)
	}

	return nil
}

// SessionPeerFind retrieves multiple SessionPeer records from the database based on the provided query.
func SessionPeerFind(db *gorm.DB, query map[string]any) ([]models.SessionPeer, error) {
	db = applyQuery(db, query)

	var peers []models.SessionPeer
	if err := db.Find(&peers).Error; err != nil {
		return nil, fmt.Errorf("finding session_peers with query %v: %w", query, err)
	}

	return peers, nil
}

// SessionPeerFindOne retrieves a single SessionPeer record from the database based on the provided query.
func SessionPeerFindOne(db *gorm.DB, query map[string]any) (*models.SessionPeer, error) {
	db = applyQuery(db, query)

	var peer models.SessionPeer
	if err := db.First(&peer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("finding session_peer with query %v: %w", query, err)
	}

	return &peer, nil
}

// SessionPeerFindOneAndUpdate finds a single SessionPeer record based on the provided query and updates it.
func SessionPeerFindOneAndUpdate(db *gorm.DB, query, updates map[string]any) (*models.SessionPeer, error) {
	var peer *models.SessionPeer
	var err error

	fn := func(db *gorm.DB) error {
		peer, err = SessionPeerFindOne(db, query)
		if err != nil {
			return fmt.Errorf("finding session_peer with query %v for update: %w", query, err)
		}

		if peer == nil {
			return nil
		}

		if err := db.Model(peer).Updates(updates).Error; err != nil {
			return fmt.Errorf("updating session_peer with query %v: %w", query, err)
		}

		return nil
	}

	if err := db.Transaction(fn); err != nil {
		return nil, fmt.Errorf("running tx: %w", err)
	}

	return peer, nil
}

// SessionPeerFindOneAndDelete finds a single SessionPeer record based on the provided query and deletes it.
func SessionPeerFindOneAndDelete(db *gorm.DB, query map[string]any) (*models.SessionPeer, error) {
	var peer *models.SessionPeer
	var err error

	fn := func(db *gorm.DB) error {
		peer, err = SessionPeerFindOne(db, query)
		if err != nil {
			return fmt.Errorf("finding session_peer with query %v for deletion: %w", query, err)
		}

		if peer == nil {
			return nil
		}

		if err := db.Model(peer).Delete(nil).Error; err != nil {
			return fmt.Errorf("deleting session_peer with query %v: %w", query, err)
		}

		return nil
	}

	if err := db.Transaction(fn); err != nil {
		return nil, fmt.Errorf("running tx: %w", err)
	}

	return peer, nil
}

// SessionAccAddrExists reports whether any session row exists for the given account address.
func SessionAccAddrExists(db *gorm.DB, addr string) (bool, error) {
	var count int64
	if err := db.Model(&models.Session{}).Where("acc_addr = ?", addr).Limit(1).Count(&count).Error; err != nil {
		return false, fmt.Errorf("checking session acc_addr %q existence: %w", addr, err)
	}

	return count > 0, nil
}

// SessionAccAddrCount returns the number of distinct account addresses across all session rows.
func SessionAccAddrCount(db *gorm.DB) (int, error) {
	var count int64
	if err := db.Model(&models.Session{}).Distinct("acc_addr").Count(&count).Error; err != nil {
		return 0, fmt.Errorf("counting distinct acc_addrs: %w", err)
	}

	return int(count), nil
}
