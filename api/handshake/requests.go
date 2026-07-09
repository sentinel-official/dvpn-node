package handshake

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	"github.com/sentinel-official/sentinel-go-sdk/v2/node"
)

// InitHandshakeRequest represents the request for performing a handshake.
type InitHandshakeRequest struct {
	Body node.InitHandshakeRequestBody
}

// NewInitHandshakeRequest parses, binds, and verifies the handshake request.
func NewInitHandshakeRequest(c *gin.Context) (req *InitHandshakeRequest, err error) {
	req = &InitHandshakeRequest{}

	// Bind JSON request to the struct.
	if err := c.ShouldBindJSON(&req.Body); err != nil {
		return nil, fmt.Errorf("binding JSON request body: %w", err)
	}

	// Verify the request body.
	if err := req.Body.Verify(); err != nil {
		return nil, fmt.Errorf("verifying request body: %w", err)
	}

	return req, nil
}

// AccAddr returns the account address from the request body.
func (r *InitHandshakeRequest) AccAddr() types.AccAddress {
	addr, err := r.Body.AccAddr()
	if err != nil {
		panic(fmt.Errorf("getting account addr from request body: %w", err))
	}

	return addr
}

// PeerRequest returns the peer request from request body.
func (r *InitHandshakeRequest) PeerRequest() []byte {
	return r.Body.Data
}
