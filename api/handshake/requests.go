package handshake

import (
	"fmt"

	cosmossdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	"github.com/sentinel-official/sentinel-go-sdk/node"
	"github.com/sentinel-official/sentinel-go-sdk/types"
)

// InitHandshakeRequest represents the request for performing a handshake.
type InitHandshakeRequest struct {
	Body node.InitHandshakeRequestBody
}

// validatePeerRequests checks the multi-element peer requests payload: the slice
// must be non-empty, each element must carry non-empty inner Data and a known
// service Type, and no Type may repeat across the slice.
func validatePeerRequests(reqs []node.PeerRequest) error {
	if len(reqs) == 0 {
		return fmt.Errorf("peer_requests cannot be empty")
	}

	seen := make(map[string]bool, len(reqs))
	for i, el := range reqs {
		if len(el.Data) == 0 {
			return fmt.Errorf("peer_requests[%d] has empty data", i)
		}

		if types.ServiceTypeFromString(el.Type) == types.ServiceTypeUnspecified {
			return fmt.Errorf("peer_requests[%d] has unknown service type %q", i, el.Type)
		}

		if seen[el.Type] {
			return fmt.Errorf("peer_requests[%d] has duplicate service type %q", i, el.Type)
		}

		seen[el.Type] = true
	}

	return nil
}

// NewInitHandshakeRequest parses, binds, validates, and verifies the handshake request.
func NewInitHandshakeRequest(c *gin.Context) (req *InitHandshakeRequest, err error) {
	req = &InitHandshakeRequest{}

	// Bind JSON request to the struct.
	if err := c.ShouldBindJSON(&req.Body); err != nil {
		return nil, fmt.Errorf("binding JSON request body: %w", err)
	}

	// Validate the peer requests shape before verifying the signature.
	if err := validatePeerRequests(req.Body.PeerRequests); err != nil {
		return nil, fmt.Errorf("validating peer requests: %w", err)
	}

	// Verify the request body.
	if err := req.Body.Verify(); err != nil {
		return nil, fmt.Errorf("verifying request body: %w", err)
	}

	return req, nil
}

// AccAddr returns the account address from the request body.
func (r *InitHandshakeRequest) AccAddr() cosmossdk.AccAddress {
	addr, err := r.Body.AccAddr()
	if err != nil {
		panic(fmt.Errorf("getting account addr from request body: %w", err))
	}

	return addr
}

// PeerRequests returns the per-protocol peer requests from the request body.
func (r *InitHandshakeRequest) PeerRequests() []node.PeerRequest {
	return r.Body.PeerRequests
}
