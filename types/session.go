package types

import (
	"time"

	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
)

type Session struct {
	ID        sdkTypes.ID        `json:"id"`
	Index     uint64             `json:"index"`
	Bandwidth sdkTypes.Bandwidth `json:"bandwidth"`
	Signature []byte             `json:"signature"`
	Status    string             `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
}
