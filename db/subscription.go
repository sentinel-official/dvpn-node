package db

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/jinzhu/gorm"

	hub "github.com/sentinel-official/hub/types"

	"github.com/sentinel-official/dvpn-node/types"
)

const (
	subscriptionTable = "subscriptions"
)

type subscription struct {
	ID        string    `json:"id" gorm:"Column:_id;Size:16;NOT NULL;PRIMARY_KEY"`
	TxHash    string    `json:"tx_hash" gorm:"Column:_tx_hash;Size:64;NOT NULL;UNIQUE"`
	Address   string    `json:"address" gorm:"Column:_address;Size:64;NOT NULL;INDEX"`
	PubKey    string    `json:"pub_key" gorm:"Column:_pub_key;Size:128;NOT NULL"`
	Upload    int64     `json:"upload" gorm:"Column:_upload;NOT NULL"`
	Download  int64     `json:"download" gorm:"Column:_download;NOT NULL"`
	Status    string    `json:"status" gorm:"Column:_status;Size:8;NOT NULL;INDEX"`
	CreatedAt time.Time `json:"created_at" gorm:"Column:_created_at;NOT NULL"`
}

func (subscription) TableName() string {
	return subscriptionTable
}

func (s *subscription) Subscription() (*types.Subscription, error) {
	address, err := sdk.AccAddressFromBech32(s.Address)
	if err != nil {
		return nil, err
	}

	pubKey, err := sdk.GetAccPubKeyBech32(s.PubKey)
	if err != nil {
		return nil, err
	}

	id, err := hub.NewSubscriptionIDFromString(s.ID)
	if err != nil {
		return nil, err
	}

	return &types.Subscription{
		ID:        id,
		TxHash:    s.TxHash,
		Address:   address,
		PubKey:    pubKey,
		Bandwidth: hub.NewBandwidthFromInt64(s.Upload, s.Download),
		Status:    s.Status,
		CreatedAt: s.CreatedAt,
	}, nil
}

func (d DB) SubscriptionSave(s *types.Subscription) error {
	pubKey, err := sdk.Bech32ifyAccPub(s.PubKey)
	if err != nil {
		return err
	}

	_s := subscription{
		ID:        s.ID.String(),
		TxHash:    s.TxHash,
		Address:   s.Address.String(),
		PubKey:    pubKey,
		Upload:    s.Bandwidth.Upload.Int64(),
		Download:  s.Bandwidth.Download.Int64(),
		Status:    s.Status,
		CreatedAt: s.CreatedAt,
	}

	return d.db.Create(&_s).Error
}

func (d DB) SubscriptionFindOne(query interface{}, args ...interface{}) (*types.Subscription, error) {
	var _s subscription

	err := d.db.Table(subscriptionTable).
		Where(query, args...).
		First(&_s).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return _s.Subscription()
}
