package utils

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidError(v error) error {
	if status.Code(v) == codes.NotFound {
		return nil
	}

	return v
}
