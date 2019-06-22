// nolint
package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ironman0x7b2/vpn-node/types"
)

var (
	testSession = types.Session{
		ID:        testIDZero,
		Index:     0,
		Bandwidth: testBandwidth1,
		Signature: []byte("signature"),
		Status:    types.ACTIVE,
		CreatedAt: testTime,
	}
	dbSession = session{
		ID:        "0",
		Index:     0,
		Upload:    1024,
		Download:  1024,
		Signature: []byte("signature"),
		Status:    types.ACTIVE,
		CreatedAt: testTime,
	}
)

func TestSession_TableName(t *testing.T) {
	name := session{}.TableName()
	require.NotEqual(t, name, "")
	require.Equal(t, name, sessionTable)
}

func TestSession_Session(t *testing.T) {
	ses, err := dbSession.Session()
	require.Nil(t, err)
	require.Equal(t, &testSession, ses)
}

func TestDB_SessionSave(t *testing.T) {
	db, err := NewDatabase(tempFile())
	require.Nil(t, err)

	query, args := "_id = ?", []interface{}{"0"}
	session0, err := db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, session0)

	err = db.SessionSave(&testSession)
	require.NotNil(t, err)

	err = db.SubscriptionSave(&testSubscription)
	require.Nil(t, err)

	err = db.SessionSave(&testSession)
	require.Nil(t, err)

	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, session0)
	require.Equal(t, &testSession, session0)

	session1 := testSession
	session1.Index++
	err = db.SessionSave(&session1)
	require.Nil(t, err)

	query, args = "_id = ? and _index = ?", []interface{}{"0", 1}
	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, session0)
	require.Equal(t, &session1, session0)
}

func TestDB_SessionFindOne(t *testing.T) {
	db, err := NewDatabase(tempFile())
	require.Nil(t, err)

	query, args := "_id = ?", []interface{}{"0"}
	session0, err := db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, session0)

	err = db.SubscriptionSave(&testSubscription)
	require.Nil(t, err)

	err = db.SessionSave(&testSession)
	require.Nil(t, err)

	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, session0)
	require.Equal(t, &testSession, session0)

	query, args = "_id = ? and _index = ?", []interface{}{"0", 0}

	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, session0)
	require.Equal(t, &testSession, session0)

	args = []interface{}{"0", 1}

	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, session0)

	query, args = "_id = ? and _upload = ? and _download = ?", []interface{}{"0", -1024, -1024}
	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, session0)

	query, args = "_id = ? and _upload = ? and _download = ?", []interface{}{"0", 1024, 1024}
	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, session0)
	require.Equal(t, &testSession, session0)

	query, args = "_id = ? and _status = ?", []interface{}{"0", types.INACTIVE}
	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, session0)

	query, args = "_id = ? and _status = ?", []interface{}{"0", types.ACTIVE}
	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, session0)
	require.Equal(t, &testSession, session0)

	query, args = "_id = ? and _created_at = ?", []interface{}{"0", testTime}
	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, session0)
	require.Equal(t, &testSession, session0)
}

func TestDB_SessionFindOneAndUpdate(t *testing.T) {
	db, err := NewDatabase(tempFile())
	require.Nil(t, err)

	query, args := "_id = ? and _index = ?", []interface{}{"0", 1}
	session0, err := db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, session0)

	err = db.SubscriptionSave(&testSubscription)
	require.Nil(t, err)

	err = db.SessionSave(&testSession)
	require.Nil(t, err)

	updates := map[string]interface{}{"_id": 1}
	query, args = "_id = ?", []interface{}{"0"}
	err = db.SessionFindOneAndUpdate(updates, query, args...)
	require.NotNil(t, err)

	query, args = "_id = ?", []interface{}{"1"}
	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.Nil(t, session0)

	subscription0 := testSubscription
	subscription0.ID = 1
	subscription0.TxHash = fmt.Sprintf("%d", 1)
	err = db.SubscriptionSave(&subscription0)
	require.Nil(t, err)

	query, args = "_id = ?", []interface{}{"0"}
	err = db.SessionFindOneAndUpdate(updates, query, args...)
	require.Nil(t, err)

	session1 := testSession
	session1.ID = 1
	query, args = "_id = ?", []interface{}{"1"}
	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, session0)
	require.Equal(t, &session1, session0)

	updates = map[string]interface{}{"_index": 1}
	err = db.SessionFindOneAndUpdate(updates, query, args...)
	require.Nil(t, err)

	session1.Index = 1
	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, session0)
	require.Equal(t, &session1, session0)

	updates = map[string]interface{}{"_upload": 2048, "_download": 2048}
	err = db.SessionFindOneAndUpdate(updates, query, args...)
	require.Nil(t, err)

	session1.Bandwidth = testBandwidth2
	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, session0)
	require.Equal(t, &session1, session0)

	updates = map[string]interface{}{"_signature": []byte("signature")}
	err = db.SessionFindOneAndUpdate(updates, query, args...)
	require.Nil(t, err)

	session1.Signature = []byte("signature")
	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, session0)
	require.Equal(t, &session1, session0)

	updates = map[string]interface{}{"_status": types.INACTIVE}
	err = db.SessionFindOneAndUpdate(updates, query, args...)
	require.Nil(t, err)

	session1.Status = types.INACTIVE
	session0, err = db.SessionFindOne(query, args...)
	require.Nil(t, err)
	require.NotNil(t, session0)
	require.Equal(t, &session1, session0)
}
