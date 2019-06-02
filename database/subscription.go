package database

import (
	"time"

	csdkTypes "github.com/cosmos/cosmos-sdk/types"

	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"

	"github.com/ironman0x7b2/vpn-node/types"
)

type subscriptionModel struct {
	ID                string    `json:"id"`
	TxHash            string    `json:"tx_hash"`
	ClientAddress     string    `json:"client_address"`
	ClientPubKey      string    `json:"client_pub_key"`
	RemainingUpload   int64     `json:"remaining_upload"`
	RemainingDownload int64     `json:"remaining_download"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}

func (subscriptionModel) TableName() string {
	return "subscriptions"
}

func (s subscriptionModel) Subscription() (*types.Subscription, error) {
	clientAddress, err := csdkTypes.AccAddressFromBech32(s.ClientAddress)
	if err != nil {
		return nil, err
	}

	clientPubKey, err := csdkTypes.GetAccPubKeyBech32(s.ClientPubKey)
	if err != nil {
		return nil, err
	}

	return &types.Subscription{
		ID:                 sdkTypes.NewIDFromString(s.ID),
		TxHash:             s.TxHash,
		ClientAddress:      clientAddress,
		ClientPubKey:       clientPubKey,
		RemainingBandwidth: sdkTypes.NewBandwidthFromInt64(s.RemainingUpload, s.RemainingDownload),
		Status:             s.Status,
		CreatedAt:          s.CreatedAt,
	}, nil
}

func (d DB) SubscriptionSave(sub *types.Subscription) error {
	clientPubKey, err := csdkTypes.Bech32ifyAccPub(sub.ClientPubKey)
	if err != nil {
		return err
	}

	_sub := subscriptionModel{
		ID:                sub.ID.String(),
		TxHash:            sub.TxHash,
		ClientAddress:     sub.ClientAddress.String(),
		ClientPubKey:      clientPubKey,
		RemainingUpload:   sub.RemainingBandwidth.Upload.Int64(),
		RemainingDownload: sub.RemainingBandwidth.Download.Int64(),
		Status:            sub.Status,
		CreatedAt:         sub.CreatedAt,
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
