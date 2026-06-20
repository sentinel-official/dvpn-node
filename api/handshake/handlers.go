package handshake

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"cosmossdk.io/math"
	cosmossdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	"github.com/sentinel-official/sentinel-go-sdk/node"
	"github.com/sentinel-official/sentinel-go-sdk/types"
	"github.com/sentinel-official/sentinelhub/v12/types/v1"
	"gorm.io/gorm"

	"github.com/sentinel-official/sentinel-dvpnx/core"
	"github.com/sentinel-official/sentinel-dvpnx/database/models"
	"github.com/sentinel-official/sentinel-dvpnx/database/operations"
)

// accountAdmitted reports whether a handshake for addr may be admitted under the
// node-wide distinct-account peer limit. An account that already holds a session
// consumes no new slot; a new account is admitted only while the distinct-account
// count is below maxPeers. Callers MUST hold the admission lock so the
// count-then-insert sequence cannot over-admit past maxPeers.
func accountAdmitted(db *gorm.DB, addr string, maxPeers uint) (bool, error) {
	exists, err := operations.SessionAccAddrExists(db, addr)
	if err != nil {
		return false, fmt.Errorf("checking account %q existence: %w", addr, err)
	}

	if exists {
		return true, nil
	}

	count, err := operations.SessionAccAddrCount(db)
	if err != nil {
		return false, fmt.Errorf("counting accounts: %w", err)
	}

	return uint(count) < maxPeers, nil
}

// handlerInitHandshake returns a handler function to process the request for performing a handshake.
func handlerInitHandshake(c *core.Context) gin.HandlerFunc { //nolint:maintidx // handler complexity is inherent in the protocol
	var mu sync.Mutex

	return func(ctx *gin.Context) {
		// Parse, validate, and verify the request (shape + signature).
		req, err := NewInitHandshakeRequest(ctx)
		if err != nil {
			err = fmt.Errorf("parsing request from context: %w", err)
			ctx.JSON(http.StatusBadRequest, types.NewResponseError(1, err))

			return
		}

		// Check if a session already exists by ID.
		query := map[string]any{"id": req.Body.ID}

		record, err := operations.SessionFindOne(c.Database(), query)
		if err != nil {
			err = fmt.Errorf("retrieving session %d from database: %w", req.Body.ID, err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(2, err))

			return
		}

		if record != nil {
			err = fmt.Errorf("session %d already exists in database", req.Body.ID)
			ctx.JSON(http.StatusConflict, types.NewResponseError(2, err))

			return
		}

		// Fetch session details from blockchain (slow; kept outside the admission lock).
		session, err := c.Client().Session(ctx, req.Body.ID)
		if err != nil {
			err = fmt.Errorf("querying session %d from blockchain: %w", req.Body.ID, err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(3, err))

			return
		}

		if session == nil {
			err = fmt.Errorf("session %d does not exist on blockchain", req.Body.ID)
			ctx.JSON(http.StatusNotFound, types.NewResponseError(3, err))

			return
		}

		// Validate session status.
		if !session.GetStatus().Equal(v1.StatusActive) {
			err = fmt.Errorf("invalid session status %q, expected %q", session.GetStatus(), v1.StatusActive)
			ctx.JSON(http.StatusBadRequest, types.NewResponseError(3, err))

			return
		}

		// Validate node address.
		if session.GetNodeAddress() != c.NodeAddr().String() {
			err = fmt.Errorf("node address mismatch: got %q, expected %q", session.GetNodeAddress(), c.NodeAddr())
			ctx.JSON(http.StatusBadRequest, types.NewResponseError(4, err))

			return
		}

		// Validate account address.
		accAddr, err := cosmossdk.AccAddressFromBech32(session.GetAccAddress())
		if err != nil {
			err = fmt.Errorf("decoding Bech32 account addr %q: %w", session.GetAccAddress(), err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(4, err))

			return
		}

		if got := req.AccAddr(); !got.Equals(accAddr) {
			err = fmt.Errorf("account addr mismatch; got %q, expected %q", got, accAddr)
			ctx.JSON(http.StatusUnauthorized, types.NewResponseError(4, err))

			return
		}

		// Per-element duplicate guard: reject if any requested peer already exists.
		// Use the canonical service-type string (same form stored by the writer) for
		// symmetry and robustness against future validation-reordering.
		for _, pr := range req.PeerRequests() {
			query := map[string]any{
				"service_type": types.ServiceTypeFromString(pr.Type).String(),
				"peer_request": base64.StdEncoding.EncodeToString(pr.Data),
			}

			peer, err := operations.SessionPeerFindOne(c.Database(), query)
			if err != nil {
				err = fmt.Errorf("retrieving session peer for %q request from database: %w", pr.Type, err)
				ctx.JSON(http.StatusInternalServerError, types.NewResponseError(5, err))

				return
			}

			if peer != nil {
				err = fmt.Errorf("session peer already exists for %q request", pr.Type)
				ctx.JSON(http.StatusConflict, types.NewResponseError(5, err))

				return
			}
		}

		// Admission, routing, and persist run under the lock so concurrent
		// handshakes cannot both pass the account-limit check and over-admit.
		mu.Lock()
		defer mu.Unlock()

		admitted, err := accountAdmitted(c.Database(), accAddr.String(), c.MaxPeers())
		if err != nil {
			err = fmt.Errorf("checking account admission for %q: %w", accAddr, err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(6, err))

			return
		}

		if !admitted {
			err = fmt.Errorf("maximum peer limit %d reached", c.MaxPeers())
			ctx.JSON(http.StatusConflict, types.NewResponseError(6, err))

			return
		}

		// Route each requested protocol to its active service, best-effort per item.
		responses := make([]node.AddPeerResponse, 0, len(req.PeerRequests()))

		peers := make([]*models.SessionPeer, 0, len(req.PeerRequests()))
		for _, pr := range req.PeerRequests() {
			t := types.ServiceTypeFromString(pr.Type)

			service, ok := c.Service(t)
			if !ok {
				responses = append(responses, node.AddPeerResponse{
					Type: pr.Type,
					Err:  "service not active",
				})

				continue
			}

			id, data, err := service.AddPeer(ctx, pr.Data)
			if err != nil {
				responses = append(responses, node.AddPeerResponse{
					Type: pr.Type,
					Err:  err.Error(),
				})

				continue
			}

			metadata, err := json.Marshal(data)
			if err != nil {
				responses = append(responses, node.AddPeerResponse{
					Type: pr.Type,
					Err:  fmt.Sprintf("encoding add-peer response: %s", err),
				})

				continue
			}

			responses = append(responses, node.AddPeerResponse{
				Type: pr.Type,
				Data: metadata,
			})

			peers = append(peers, models.NewSessionPeer().
				WithSessionID(req.Body.ID).
				WithServiceType(t).
				WithPeerID(id).
				WithPeerRequest(pr.Data).
				WithPeerMetadata(metadata).
				WithRxBytes(math.ZeroInt()).
				WithTxBytes(math.ZeroInt()).
				WithDuration(0))
		}

		// Persist the parent session and its successful child peers (parent first
		// so the child foreign keys resolve).
		if len(peers) > 0 {
			item := models.NewSession().
				WithID(session.GetID()).
				WithNodeAddr(c.NodeAddr()).
				WithAccAddr(accAddr).
				WithMaxBytes(session.GetMaxBytes()).
				WithMaxDuration(session.GetMaxDuration()).
				WithSignature(nil)

			if err = operations.SessionInsertOne(c.Database(), item); err != nil {
				err = fmt.Errorf("inserting session %d into database: %w", item.GetID(), err)
				ctx.JSON(http.StatusInternalServerError, types.NewResponseError(7, err))

				return
			}

			for _, peer := range peers {
				if err = operations.SessionPeerInsertOne(c.Database(), peer); err != nil {
					err = fmt.Errorf("inserting session peer for session %d into database: %w", req.Body.ID, err)
					ctx.JSON(http.StatusInternalServerError, types.NewResponseError(7, err))

					return
				}
			}
		}

		// Return a successful response with per-item results.
		res := &node.InitHandshakeResult{
			Addrs:            c.RemoteAddrs(),
			AddPeerResponses: responses,
		}

		ctx.JSON(http.StatusOK, types.NewResponseResult(res))
	}
}
