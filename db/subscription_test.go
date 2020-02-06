// nolint
package db

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	hub "github.com/sentinel-official/hub/types"

	"github.com/sentinel-official/dvpn-node/types"
)

var (
	testPubKey          = ed25519.GenPrivKey().PubKey()
	testBech32PubKey, _ = sdk.Bech32ifyAccPub(testPubKey)
	testAddress         = sdk.AccAddress(testPubKey.Address())
	testSessIDZero, _   = hub.NewSessionIDFromString("sess0")
	testSessIDOne, _    = hub.NewSessionIDFromString("sess1")
	testSubsIDZero, _   = hub.NewSubscriptionIDFromString("subs0")
	testSubsIDOne, _    = hub.NewSubscriptionIDFromString("subs1")
	testBandwidth1      = hub.NewBandwidthFromInt64(1024, 1024)
	testBandwidth2      = hub.NewBandwidthFromInt64(2048, 2048)
	testTime            = time.Now().UTC()
)

var (
	testSubscription = types.Subscription{
		ID:        testSubsIDZero,
		TxHash:    fmt.Sprintf("%d", 0),
		Address:   testAddress,
		PubKey:    testPubKey,
		Bandwidth: testBandwidth1,
		Status:    types.ACTIVE,
		CreatedAt: testTime,
	}
	testDBSubscription = subscription{
		ID:        "0",
		TxHash:    fmt.Sprintf("%d", 0),
		Address:   testAddress.String(),
		PubKey:    testBech32PubKey,
		Upload:    1024,
		Download:  1024,
		Status:    types.ACTIVE,
		CreatedAt: testTime,
	}
)

func tempFile() string {
	file, err := ioutil.TempFile(os.TempDir(), "db_")
	if err != nil {
		panic(err)
	}

	return file.Name()
}

func TestSubscription_TableName(t *testing.T) {
	name := subscription{}.TableName()
	require.NotEqual(t, name, "")
	require.Equal(t, name, subscriptionTable)
}

func TestSubscription_Subscription(t *testing.T) {
	subscription0 := testDBSubscription
	sub, err := subscription0.Subscription()
	require.Nil(t, err)
	require.NotNil(t, sub)
	require.Equal(t, &testSubscription, sub)

	subscription0.Address = "address"
	sub, err = subscription0.Subscription()
	require.NotNil(t, err)
	require.Nil(t, sub)

	subscription0.Address = ""
	sub, err = subscription0.Subscription()
	require.Nil(t, err)
	require.NotNil(t, sub)

	subscription0.PubKey = "pub_key"
	sub, err = subscription0.Subscription()
	require.NotNil(t, err)
	require.Nil(t, sub)

	subscription0.PubKey = ""
	sub, err = subscription0.Subscription()
	require.NotNil(t, err)
	require.Nil(t, sub)
}

func TestDB_SubscriptionSave(t *testing.T) {
	db, err := NewDatabase(tempFile())
	require.Nil(t, err)

	query, args := "_id = ?", []interface{}{"0"}
	subscription0, err := db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, subscription0)

	err = db.SubscriptionSave(&testSubscription)
	require.Nil(t, err)

	subscription0, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, subscription0)
	require.Equal(t, &testSubscription, subscription0)

	subscription0, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, subscription0)
	require.Equal(t, &testSubscription, subscription0)

	err = db.SubscriptionSave(&testSubscription)
	require.NotNil(t, err)

	subscription1 := testSubscription
	subscription1.ID = testSubsIDOne
	subscription1.TxHash = fmt.Sprintf("%d", 1)
	err = db.SubscriptionSave(&subscription1)
	require.Nil(t, err)

	query, args = "_id = ?", []interface{}{"1"}
	subscription0, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, subscription0)
	require.Equal(t, &subscription1, subscription0)
}

func TestDB_SubscriptionFindOne(t *testing.T) {
	db, err := NewDatabase(tempFile())
	require.Nil(t, err)

	query, args := "_id = ?", []interface{}{"0"}
	subscription0, err := db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, subscription0)

	err = db.SubscriptionSave(&testSubscription)
	require.Nil(t, err)

	subscription0, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, subscription0)
	require.Equal(t, &testSubscription, subscription0)

	args = []interface{}{"1"}
	subscription0, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, subscription0)

	query, args = "_id = ? and _tx_hash = ?", []interface{}{"0", fmt.Sprintf("%d", 1)}
	subscription0, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, subscription0)

	args = []interface{}{"0", fmt.Sprintf("%d", 0)}
	subscription0, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, subscription0)
	require.Equal(t, &testSubscription, subscription0)

	query, args = "_id = ? and _address = ?", []interface{}{"0", testAddress.String()}
	subscription0, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, subscription0)
	require.Equal(t, &testSubscription, subscription0)

	query, args = "_id = ? and _pub_key = ?", []interface{}{"0", testBech32PubKey}
	subscription0, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, subscription0)
	require.Equal(t, &testSubscription, subscription0)

	query, args = "_id = ? and _upload = ? and _download = ?", []interface{}{"0", -1024, -1024}
	subscription0, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, subscription0)

	args = []interface{}{"0", 1024, 1024}
	subscription0, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, subscription0)
	require.Equal(t, &testSubscription, subscription0)

	query, args = "_id = ? and _status = ?", []interface{}{"0", types.INACTIVE}
	subscription0, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, subscription0)

	args = []interface{}{"0", types.ACTIVE}
	subscription0, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, subscription0)
	require.Equal(t, &testSubscription, subscription0)

	query, args = "_id = ? and _created_at = ?", []interface{}{"0", testTime}
	subscription0, err = db.SubscriptionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, subscription0)
	require.Equal(t, &testSubscription, subscription0)
}
