package context

import (
	"encoding/base64"
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

	return nil
}
