package types

import (
	"time"

	csdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/ironman0x7b2/sentinel-sdk/types"
)

type Subscription struct {
	ID        sdk.ID         `json:"id"`
	TxHash    string          `json:"tx_hash"`
	Address   csdk.AccAddress `json:"address"`
	PubKey    crypto.PubKey   `json:"pub_key"`
	Bandwidth sdk.Bandwidth  `json:"bandwidth"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
}
