package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gorm.io/gorm"
)

type Session struct {
	gorm.Model
	ID           uint64 `gorm:"primaryKey;uniqueIndex:idx_sessions_id"`
	Subscription uint64 `gorm:"index:idx_sessions_subscription_address"`
	Key          string `gorm:"uniqueIndex:idx_sessions_key"`
	Address      string `gorm:"index:idx_sessions_address;index:idx_sessions_subscription_address"`
	Available    int64
	Download     int64
	Upload       int64
}

func (s *Session) GetAddress() sdk.AccAddress {
	if s.Address == "" {
		return nil
	}

	address, err := sdk.AccAddressFromBech32(s.Address)
	if err != nil {
		panic(err)
	}

	return address
}
