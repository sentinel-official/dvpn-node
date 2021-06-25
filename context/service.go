package context

import (
	"encoding/base64"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (c *Context) RemovePeer(key string) error {
	c.Log().Info("Removing peer from service", "key", key)

	data, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		c.Log().Error("Failed to decode the key", "error", err)
		return err
	}

	if err := c.Service().RemovePeer(data); err != nil {
		c.Log().Error("Failed to remove the peer from service", "error", err)
		return err
	}

	c.Log().Info("Removed peer from service...")
	return nil
}

func (c *Context) RemoveSession(key string, address sdk.AccAddress) error {
	c.Log().Info("Removing session from list", "key", key, "address", address)

	c.Sessions().DeleteByKey(key)
	c.Sessions().DeleteByAddress(address)

	c.Log().Info("Removed session from list...")
	return nil
}

func (c *Context) RemovePeerAndSession(key string, address sdk.AccAddress) error {
	if err := c.RemovePeer(key); err != nil {
		return err
	}
	if err := c.RemoveSession(key, address); err != nil {
		return err
	}

	return nil
}
