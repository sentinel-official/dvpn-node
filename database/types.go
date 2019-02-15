package database

import (
	"time"

	"github.com/ironman0x7b2/sentinel-sdk/x/vpn"
	"gopkg.in/mgo.v2/bson"
)

type Session struct {
	ID bson.ObjectId `json:"_id"`

	SessionID      string          `json:"session_id"`
	ClientAddress  string          `json:"client_address"`
	UploadLimit    int64           `json:"upload_limit"`
	DownloadLimit  int64           `json:"download_limit"`
	PricePerGB     string          `json:"price_per_gb"`
	BandwidthSigns []BandwidthSign `json:"bandwidth_signs"`

	Status    string    `json:"status"`
	AddedAt   time.Time `json:"added_at"`
	StatusAt  time.Time `json:"status_at"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
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
	Index         int64     `json:"index"`
	Upload        int64     `json:"upload"`
	Download      int64     `json:"download"`
	ClientSign    string    `json:"client_sign"`
	NodeOwnerSign string    `json:"node_owner_sign"`
	TxHash        string    `json:"tx_hash"`
	AddedAt       time.Time `json:"added_at"`
}
