package types

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func QueryError(v error) error {
	if status.Code(v) == codes.NotFound {
		return nil
	}

	return v
}
