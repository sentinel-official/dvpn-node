package database

import (
	"time"

	csdkTypes "github.com/cosmos/cosmos-sdk/types"

	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"

	"github.com/ironman0x7b2/vpn-node/types"
)

type subscriptionModel struct {
	ID        string    `json:"id"`
	TxHash    string    `json:"tx_hash"`
	Address   string    `json:"address"`
	PubKey    string    `json:"pub_key"`
	Upload    int64     `json:"upload"`
	Download  int64     `json:"download"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (subscriptionModel) TableName() string {
	return "subscriptions"
}

func (s subscriptionModel) Subscription() (*types.Subscription, error) {
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

func (d DB) SubscriptionSave(sub *types.Subscription) error {
	pubKey, err := csdkTypes.Bech32ifyAccPub(sub.PubKey)
	if err != nil {
		return err
	}

	_sub := subscriptionModel{
		ID:        sub.ID.String(),
		TxHash:    sub.TxHash,
		Address:   sub.Address.String(),
		PubKey:    pubKey,
		Upload:    sub.Bandwidth.Upload.Int64(),
		Download:  sub.Bandwidth.Download.Int64(),
		Status:    sub.Status,
		CreatedAt: sub.CreatedAt,
	}

	d.db.Create(&_sub)
	return nil
}

func (d DB) SubscriptionFindOne(query interface{}, args ...interface{}) (*types.Subscription, error) {
	var _sub subscriptionModel
	d.db.Table("subscriptions").Where(query, args...).First(&_sub)

	sub, err := _sub.Subscription()
	if err != nil {
		return nil, err
	}

	return sub, nil
}
