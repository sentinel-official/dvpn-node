package handshake

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"cosmossdk.io/math"
	cosmossdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	"github.com/sentinel-official/sentinel-go-sdk/node"
	"github.com/sentinel-official/sentinel-go-sdk/types"
	"github.com/sentinel-official/sentinelhub/v12/types/v1"

	"github.com/sentinel-official/sentinel-dvpnx/core"
	"github.com/sentinel-official/sentinel-dvpnx/database/models"
	"github.com/sentinel-official/sentinel-dvpnx/database/operations"
)

// handlerInitHandshake returns a handler function to process the request for performing a handshake.
func handlerInitHandshake(c *core.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Reject handshake if maximum peer limit is reached
		if n := c.Service().PeersLen(); uint(n) >= c.MaxPeers() {
			err := fmt.Errorf("maximum peer limit %d reached", n)
			ctx.JSON(http.StatusConflict, types.NewResponseError(1, err))

			return
		}

		// Parse and verify the request.
		req, err := NewInitHandshakeRequest(ctx)
		if err != nil {
			err = fmt.Errorf("parsing request from context: %w", err)
			ctx.JSON(http.StatusBadRequest, types.NewResponseError(2, err))

			return
		}

		// Check if a session already exists by ID.
		query := map[string]any{
			"id": req.Body.ID,
		}

		record, err := operations.SessionFindOne(c.Database(), query)
		if err != nil {
			err = fmt.Errorf("retrieving session %d from database: %w", req.Body.ID, err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(3, err))

			return
		}

		if record != nil {
			err = fmt.Errorf("session %d already exists in database", req.Body.ID)
			ctx.JSON(http.StatusConflict, types.NewResponseError(3, err))

			return
		}

		// Check if a session already exists by peer request data.
		peerReqStr := base64.StdEncoding.EncodeToString(req.PeerRequest())
		query = map[string]any{
			"peer_request": peerReqStr,
		}

		record, err = operations.SessionFindOne(c.Database(), query)
		if err != nil {
			err = fmt.Errorf("retrieving session for peer request %q from database: %w", peerReqStr, err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(4, err))

			return
		}

		if record != nil {
			err = fmt.Errorf("session already exists for peer request %q", peerReqStr)
			ctx.JSON(http.StatusConflict, types.NewResponseError(4, err))

			return
		}

		// Fetch session details from blockchain.
		session, err := c.Client().Session(ctx, req.Body.ID)
		if err != nil {
			err = fmt.Errorf("querying session %d from blockchain: %w", req.Body.ID, err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(5, err))

			return
		}

		if session == nil {
			err = fmt.Errorf("session %d does not exist on blockchain", req.Body.ID)
			ctx.JSON(http.StatusNotFound, types.NewResponseError(5, err))

			return
		}

		// Validate session status.
		if !session.GetStatus().Equal(v1.StatusActive) {
			err = fmt.Errorf("invalid session status %q, expected %q", session.GetStatus(), v1.StatusActive)
			ctx.JSON(http.StatusBadRequest, types.NewResponseError(5, err))

			return
		}

		// Validate node address.
		if session.GetNodeAddress() != c.NodeAddr().String() {
			err = fmt.Errorf("node address mismatch: got %q, expected %q", session.GetNodeAddress(), c.NodeAddr())
			ctx.JSON(http.StatusBadRequest, types.NewResponseError(6, err))

			return
		}

		// Validate account address.
		accAddr, err := cosmossdk.AccAddressFromBech32(session.GetAccAddress())
		if err != nil {
			err = fmt.Errorf("decoding Bech32 account addr %q: %w", session.GetAccAddress(), err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(6, err))

			return
		}

		if got := req.AccAddr(); !got.Equals(accAddr) {
			err = fmt.Errorf("account addr mismatch; got %q, expected %q", got, accAddr)
			ctx.JSON(http.StatusUnauthorized, types.NewResponseError(6, err))

			return
		}

		// Add the peer to the active service.
		id, data, err := c.Service().AddPeer(ctx, req.PeerRequest())
		if err != nil {
			err = fmt.Errorf("adding peer to service: %w", err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(7, err))

			return
		}

		// Encode and prepare the handshake response.
		res := &node.InitHandshakeResult{Addrs: c.RemoteAddrs()}
		if res.Data, err = json.Marshal(data); err != nil {
			err = fmt.Errorf("encoding add-peer service response: %w", err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(8, err))

			return
		}

		// Insert the session record into the database.
		item := models.NewSession().
			WithAccAddr(accAddr).
			WithDuration(0).
			WithID(session.GetID()).
			WithMaxBytes(session.GetMaxBytes()).
			WithMaxDuration(session.GetMaxDuration()).
			WithNodeAddr(c.NodeAddr()).
			WithPeerID(id).
			WithPeerMetadata(res.Data).
			WithPeerRequest(req.PeerRequest()).
			WithRxBytes(math.ZeroInt()).
			WithServiceType(c.Service().Type()).
			WithSignature(nil).
			WithTxBytes(math.ZeroInt())

		if err = operations.SessionInsertOne(c.Database(), item); err != nil {
			err = fmt.Errorf("inserting session %d into database: %w", item.GetID(), err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(9, err))

			return
		}

		// Return a successful response.
		ctx.JSON(http.StatusOK, types.NewResponseResult(res))
	}
}
