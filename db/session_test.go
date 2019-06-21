package db

import (
	"fmt"
	"syscall"
	"testing"
	"time"

	csdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/ironman0x7b2/sentinel-sdk/types"

	"github.com/ironman0x7b2/vpn-node/types"
)

func TestDB_SessionSave(t *testing.T) {
	_ = syscall.Unlink("/tmp/sentinel.db")

	db, err := NewDatabase("/tmp/sentinel.db")
	require.NotNil(t, db)
	require.Nil(t, err)

	query, args := "_id = ?", []interface{}{
		"0",
	}

	_ses, err := db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, _ses)

	ses := &types.Session{
		ID:        sdk.NewIDFromString("0"),
		Index:     0,
		Bandwidth: sdk.NewBandwidthFromInt64(1024, 1024),
		Signature: nil,
		Status:    types.ACTIVE,
		CreatedAt: time.Now().UTC(),
	}

	err = db.SessionSave(ses)
	require.NotNil(t, err)

	_ses, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, _ses)

	pubKey := ed25519.GenPrivKey().PubKey()
	address := csdk.AccAddress(pubKey.Address())
	sub := &types.Subscription{
		ID:        sdk.NewIDFromUInt64(0),
		TxHash:    fmt.Sprintf("%d", 0),
		Address:   address,
		PubKey:    pubKey,
		Bandwidth: sdk.NewBandwidthFromInt64(1024, 1024),
		Status:    types.ACTIVE,
		CreatedAt: time.Now().UTC(),
	}

	err = db.SubscriptionSave(sub)
	require.Nil(t, err)

	err = db.SessionSave(ses)
	require.Nil(t, err)

	_ses, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, _ses)
	require.Equal(t, ses, _ses)
}
