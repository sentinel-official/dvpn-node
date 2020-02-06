package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"

	hub "github.com/sentinel-official/hub/types"
)

type Subscription struct {
	ID         hub.SubscriptionID `json:"id"`
	ResolverID hub.ResolverID     `json:"resolver_id"`
	TxHash     string             `json:"tx_hash"`
	Address    sdk.AccAddress     `json:"address"`
	PubKey     crypto.PubKey      `json:"pub_key"`
	Bandwidth  hub.Bandwidth      `json:"bandwidth"`
	Status     string             `json:"status"`
	CreatedAt  time.Time          `json:"created_at"`
}
