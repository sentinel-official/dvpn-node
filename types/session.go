package types

import (
	"time"

	sdk "github.com/ironman0x7b2/sentinel-sdk/types"
)

type Session struct {
	ID        sdk.ID        `json:"id"`
	Index     uint64        `json:"index"`
	Bandwidth sdk.Bandwidth `json:"bandwidth"`
	Signature []byte        `json:"signature"`
	Status    string        `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
}
