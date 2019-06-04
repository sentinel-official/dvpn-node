package database

import (
	"time"

	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"

	"github.com/ironman0x7b2/vpn-node/types"
)

type session struct {
	ID        string    `json:"id" gorm:"Column:_id;Size:16;NOT NULL;PRIMARY_KEY"`
	Index     uint64    `json:"index" gorm:"Column:_index;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT:false"`
	Upload    int64     `json:"upload" gorm:"Column:_upload;NOT NULL"`
	Download  int64     `json:"download" gorm:"Column:_download;NOT NULL"`
	Signature []byte    `json:"signature" gorm:"Column:_signature"`
	Status    string    `json:"status" gorm:"Column:_status;Size:8;NOT NULL"`
	CreatedAt time.Time `json:"created_at" gorm:"Column:_created_at;NOT NULL"`
}

func (session) TableName() string {
	return "sessions"
}

func (s *session) Session() (*types.Session, error) {
	return &types.Session{
		ID:        sdkTypes.NewIDFromString(s.ID),
		Index:     s.Index,
		Bandwidth: sdkTypes.NewBandwidthFromInt64(s.Upload, s.Download),
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

	return d.db.Table("sessions").Create(&_s).Error
}

func (d DB) SessionFindOne(query interface{}, args ...interface{}) (*types.Session, error) {
	var _s session

	result := d.db.Table("sessions").Where(query, args...).First(&_s)
	if result.Error != nil {
		return nil, result.Error
	}

	session, err := _s.Session()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (d DB) SessionFindOneAndUpdate(values map[string]interface{}, query interface{}, args ...interface{}) error {
	return d.db.Table("sessions").Where(query, args...).Updates(values).Error
}
