package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/rest/session"
	"github.com/sentinel-official/dvpn-node/rest/status"
)

func RegisterRoutes(ctx *context.Context, r gin.IRouter) {
	session.RegisterRoutes(ctx, r)
	status.RegisterRoutes(ctx, r)
}
