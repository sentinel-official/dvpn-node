package api

import (
	"github.com/gin-gonic/gin"

	"github.com/sentinel-official/dvpn-node/api/session"
	"github.com/sentinel-official/dvpn-node/api/status"
	"github.com/sentinel-official/dvpn-node/context"
)

func RegisterRoutes(ctx *context.Context, r gin.IRouter) {
	session.RegisterRoutes(ctx, r)
	status.RegisterRoutes(ctx, r)
}
