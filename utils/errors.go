package utils

import (
	"strings"
)

func IsNotFoundError(v error) error {
	if strings.Contains(v.Error(), "code = NotFound") {
		return nil
	}

	return v
}
