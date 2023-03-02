package context

import (
	"encoding/base64"
)

func (c *Context) RemovePeer(key string) error {
	c.Log().Info("Removing the peer from service", "key", key)

	data, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		c.Log().Error("failed to decode the key", "error", err, "key", key)
		return err
	}

	if err = c.Service().RemovePeer(data); err != nil {
		c.Log().Error("failed to remove the peer from service", "error", err, "data", data)
		return err
	}

	return nil
}

func (c *Context) HasPeer(key string) (bool, error) {
	data, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		c.Log().Error("failed to decode the key", "error", err, "key", key)
		return false, err
	}

	return c.Service().HasPeer(data), nil
}

func (c *Context) RemovePeerIfExists(key string) error {
	ok, err := c.HasPeer(key)
	if err != nil {
		return err
	}
	if !ok {
		c.Log().Debug("Peer does not exist", "key", key)
		return nil
	}

	return c.RemovePeer(key)
}
