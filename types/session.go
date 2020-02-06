package types

import (
	"time"

	hub "github.com/sentinel-official/hub/types"
)

type Session struct {
	ID        hub.SubscriptionID `json:"id"`
	Index     uint64             `json:"index"`
	Bandwidth hub.Bandwidth      `json:"bandwidth"`
	Signature []byte             `json:"signature"`
	Status    string             `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
}
