package database

import (
	"time"

	"github.com/ironman0x7b2/sentinel-sdk/x/vpn"
	"gopkg.in/mgo.v2/bson"
)

type Session struct {
	ID bson.ObjectId `json:"id" bson:"_id"`

	SessionID      string          `json:"session_id" bson:"sessionID"`
	ClientAddress  string          `json:"client_address" bson:"clientAddress"`
	UploadLimit    int64           `json:"upload_limit" bson:"uploadLimit"`
	DownloadLimit  int64           `json:"download_limit" bson:"downloadLimit"`
	PricePerGB     string          `json:"price_per_gb" bson:"pricePerGB"`
	BandwidthSigns []BandwidthSign `json:"bandwidth_signs" bson:"bandwidthSigns"`

	Status    string    `json:"status" bson:"status"`
	AddedAt   time.Time `json:"added_at" bson:"addedAt"`
	StatusAt  time.Time `json:"status_at" bson:"statusAt"`
	StartedAt time.Time `json:"started_at" bson:"startedAt,omitempty"`
	EndedAt   time.Time `json:"ended_at" bson:"endedAt,omitempty"`
}

func NewSessionFromDetails(details *vpn.SessionDetails) Session {
	now := time.Now().UTC()

	return Session{
		ID:             bson.NewObjectId(),
		SessionID:      details.ID.String(),
		ClientAddress:  details.Client.String(),
		UploadLimit:    details.Bandwidth.ToProvide.Upload.Int64(),
		DownloadLimit:  details.Bandwidth.ToProvide.Download.Int64(),
		PricePerGB:     details.PricePerGB.String(),
		BandwidthSigns: []BandwidthSign{},
		Status:         details.Status,
		AddedAt:        now,
		StatusAt:       now,
	}
}

type BandwidthSign struct {
	Index         int64     `json:"index" bson:"index"`
	Upload        int64     `json:"upload" bson:"upload"`
	Download      int64     `json:"download" bson:"download"`
	ClientSign    string    `json:"client_sign" bson:"clientSign,omitempty"`
	NodeOwnerSign string    `json:"node_owner_sign" bson:"nodeOwnerSign"`
	TxHash        string    `json:"tx_hash" bson:"txHash,omitempty"`
	AddedAt       time.Time `json:"added_at" bson:"addedAt"`
}

func NewBandwidthSign(index, upload, download int64, sign string) BandwidthSign {
	now := time.Now().UTC()

	return BandwidthSign{
		Index:         index,
		Upload:        upload,
		Download:      download,
		NodeOwnerSign: sign,
		AddedAt:       now,
	}
}
