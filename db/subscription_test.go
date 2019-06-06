package db

import (
	"fmt"
	"syscall"
	"testing"
	"time"

	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/ironman0x7b2/vpn-node/types"
)

func TestDB_SubscriptionSave(t *testing.T) {
	_ = syscall.Unlink("/tmp/sentinel.db")

	db, err := NewDatabase("/tmp/sentinel.db")
	require.NotNil(t, db)
	require.Nil(t, err)

	query, args := "_id = ?", []interface{}{
		"0",
	}

	_sub, err := db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, _sub)

	pubKey := ed25519.GenPrivKey().PubKey()
	address := csdkTypes.AccAddress(pubKey.Address())
	sub := &types.Subscription{
		ID:        sdkTypes.NewIDFromUInt64(0),
		TxHash:    fmt.Sprintf("%d", 0),
		Address:   address,
		PubKey:    pubKey,
		Bandwidth: sdkTypes.NewBandwidthFromInt64(1024, 1024),
		Status:    types.ACTIVE,
		CreatedAt: time.Now().UTC(),
	}

	err = db.SubscriptionSave(sub)
	require.Nil(t, err)

	_sub, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, _sub)
	require.Equal(t, sub, _sub)
}
