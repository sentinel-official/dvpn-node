package database

import (
	"time"

	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"

	"github.com/ironman0x7b2/vpn-node/types"
)

type subscription struct {
	ID        string    `json:"id" gorm:"Column:_id;Size:16;NOT NULL;PRIMARY_KEY"`
	TxHash    string    `json:"tx_hash" gorm:"Column:_tx_hash;Size:64;NOT NULL;UNIQUE"`
	Address   string    `json:"address" gorm:"Column:_address;Size:64;NOT NULL"`
	PubKey    string    `json:"pub_key" gorm:"Column:_pub_key;Size:128;NOT NULL"`
	Upload    int64     `json:"upload" gorm:"Column:_upload;NOT NULL"`
	Download  int64     `json:"download" gorm:"Column:_download;NOT NULL"`
	Status    string    `json:"status" gorm:"Column:_status;Size:8;NOT NULL"`
	CreatedAt time.Time `json:"created_at" gorm:"Column:_created_at;NOT NULL"`
}

func (subscription) TableName() string {
	return "subscriptions"
}

func (s *subscription) Subscription() (*types.Subscription, error) {
	address, err := csdkTypes.AccAddressFromBech32(s.Address)
	if err != nil {
		return nil, err
	}

	pubKey, err := csdkTypes.GetAccPubKeyBech32(s.PubKey)
	if err != nil {
		return nil, err
	}

	return &types.Subscription{
		ID:        sdkTypes.NewIDFromString(s.ID),
		TxHash:    s.TxHash,
		Address:   address,
		PubKey:    pubKey,
		Bandwidth: sdkTypes.NewBandwidthFromInt64(s.Upload, s.Download),
		Status:    s.Status,
		CreatedAt: s.CreatedAt,
	}, nil
}

func (d DB) SubscriptionSave(s *types.Subscription) error {
	pubKey, err := csdkTypes.Bech32ifyAccPub(s.PubKey)
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

	result := d.db.Table("subscriptions").Where(query, args...).First(&_s)
	if result.Error != nil {
		return nil, result.Error
	}

	sub, err := _s.Subscription()
	if err != nil {
		return nil, err
	}

	return sub, nil
}
