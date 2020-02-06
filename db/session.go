package db

import (
	"time"

	"github.com/jinzhu/gorm"

	hub "github.com/sentinel-official/hub/types"

	"github.com/sentinel-official/dvpn-node/types"
)

const (
	sessionTable = "sessions"
)

type session struct {
	ID        string    `json:"id" gorm:"Column:_id;type:string REFERENCES subscriptions(_id) ON DELETE CASCADE ON UPDATE CASCADE;Size:16;PRIMARY_KEY"` // nolint:lll
	Index     uint64    `json:"index" gorm:"Column:_index;PRIMARY_KEY;AUTO_INCREMENT:false;DEFAULT:0"`
	Upload    int64     `json:"upload" gorm:"Column:_upload;NOT NULL"`
	Download  int64     `json:"download" gorm:"Column:_download;NOT NULL"`
	Signature []byte    `json:"signature" gorm:"Column:_signature"`
	Status    string    `json:"status" gorm:"Column:_status;Size:8;NOT NULL;INDEX"`
	CreatedAt time.Time `json:"created_at" gorm:"Column:_created_at;NOT NULL"`
}

func (session) TableName() string {
	return sessionTable
}

func (s *session) Session() (*types.Session, error) {
	id, err := hub.NewSubscriptionIDFromString(s.ID)
	if err != nil {
		return &types.Session{}, err
	}

	return &types.Session{
		ID:        id,
		Index:     s.Index,
		Bandwidth: hub.NewBandwidthFromInt64(s.Upload, s.Download),
		Signature: s.Signature,
		Status:    s.Status,
		CreatedAt: s.CreatedAt,
	}, nil
}

func (d DB) SessionSave(s *types.Session) error {
	_s := session{
		ID:        s.ID.String(),
		Index:     s.Index,
		Upload:    s.Bandwidth.Upload.Int64(),
		Download:  s.Bandwidth.Download.Int64(),
		Signature: s.Signature,
		Status:    s.Status,
		CreatedAt: s.CreatedAt,
	}

	return d.db.Table(sessionTable).Create(&_s).Error
}

func (d DB) SessionFindOne(query interface{}, args ...interface{}) (*types.Session, error) {
	var _s session

	err := d.db.Table(sessionTable).
		Where(query, args...).
		First(&_s).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return _s.Session()
}

func (d DB) SessionFindOneAndUpdate(values map[string]interface{}, query interface{}, args ...interface{}) error {
	return d.db.Table(sessionTable).
		Where(query, args...).
		Updates(values).Error
}
