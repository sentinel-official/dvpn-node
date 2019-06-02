package database

import (
	"time"

	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"

	"github.com/ironman0x7b2/vpn-node/types"
)

type sessionModel struct {
	ID        string    `json:"id"`
	Index     uint64    `json:"index"`
	Upload    int64     `json:"upload"`
	Download  int64     `json:"download"`
	Signature []byte    `json:"signature"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (sessionModel) TableName() string {
	return "sessions"
}

func (s sessionModel) Session() (*types.Session, error) {
	return &types.Session{
		ID:        sdkTypes.NewIDFromString(s.ID),
		Index:     s.Index,
		Bandwidth: sdkTypes.NewBandwidthFromInt64(s.Upload, s.Download),
		Signature: s.Signature,
		Status:    s.Status,
		CreatedAt: s.CreatedAt,
	}, nil
}

func (d DB) SessionSave(session *types.Session) error {
	_session := sessionModel{
		ID:        session.ID.String(),
		Index:     session.Index,
		Upload:    session.Bandwidth.Upload.Int64(),
		Download:  session.Bandwidth.Download.Int64(),
		Signature: session.Signature,
		Status:    session.Status,
		CreatedAt: session.CreatedAt,
	}

	d.db.Table("sessions").Create(&_session)
	return nil
}

func (d DB) SessionFindOne(query interface{}, args ...interface{}) (*types.Session, error) {
	var _session sessionModel
	d.db.Table("sessions").Where(query, args...).First(&_session)

	session, err := _session.Session()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (d DB) SessionFindOneAndUpdate(values map[string]interface{}, query interface{}, args ...interface{}) error {
	d.db.Table("sessions").Where(query, args...).Updates(values)
	return nil
}
