package context

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	hub "github.com/sentinel-official/hub/types"
	"github.com/sentinel-official/hub/x/node"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/sentinel-official/dvpn-node/lite"
	"github.com/sentinel-official/dvpn-node/types"
)

type Context struct {
	home      string
	ctx       context.Context
	logger    log.Logger
	service   types.Service
	client    *lite.Client
	config    *types.Config
	router    *mux.Router
	sessions  *types.Sessions
	bandwidth hub.Bandwidth
}

func NewContext() *Context {
	return &Context{
		ctx: context.Background(),
	}
}

func (c *Context) WithBandwidth(v hub.Bandwidth) *Context  { c.bandwidth = v; return c }
func (c *Context) WithClient(v *lite.Client) *Context      { c.client = v; return c }
func (c *Context) WithConfig(v *types.Config) *Context     { c.config = v; return c }
func (c *Context) WithContext(v context.Context) *Context  { c.ctx = v; return c }
func (c *Context) WithHome(v string) *Context              { c.home = v; return c }
func (c *Context) WithLogger(v log.Logger) *Context        { c.logger = v; return c }
func (c *Context) WithRouter(v *mux.Router) *Context       { c.router = v; return c }
func (c *Context) WithService(v types.Service) *Context    { c.service = v; return c }
func (c *Context) WithSessions(v *types.Sessions) *Context { c.sessions = v; return c }

func (c *Context) WithValue(key, value interface{}) *Context {
	c.WithContext(context.WithValue(c.ctx, key, value))
	return c
}

func (c *Context) Address() hub.NodeAddress          { return c.Operator().Bytes() }
func (c *Context) Bandwidth() hub.Bandwidth          { return c.bandwidth }
func (c *Context) Type() node.Category               { return c.service.Type() }
func (c *Context) Client() *lite.Client              { return c.client }
func (c *Context) Config() *types.Config             { return c.config }
func (c *Context) Context() context.Context          { return c.ctx }
func (c *Context) Home() string                      { return c.home }
func (c *Context) ListenOn() string                  { return c.Config().Node.ListenOn }
func (c *Context) Logger() log.Logger                { return c.logger }
func (c *Context) Moniker() string                   { return c.Config().Node.Moniker }
func (c *Context) Operator() sdk.AccAddress          { return c.client.FromAddress() }
func (c *Context) RemoteURL() string                 { return c.Config().Node.RemoteURL }
func (c *Context) Router() *mux.Router               { return c.router }
func (c *Context) Service() types.Service            { return c.service }
func (c *Context) Sessions() *types.Sessions         { return c.sessions }
func (c *Context) Value(key interface{}) interface{} { return c.ctx.Value(key) }
func (c *Context) Version() string                   { return types.Version }

func (c *Context) Provider() hub.ProvAddress {
	if c.Config().Node.Provider == "" {
		return nil
	}

	address, err := hub.ProvAddressFromBech32(c.Config().Node.Provider)
	if err != nil {
		panic(err)
	}

	return address
}

func (c *Context) Price() sdk.Coins {
	if c.Config().Node.Price == "" {
		return nil
	}

	coins, err := sdk.ParseCoins(c.Config().Node.Price)
	if err != nil {
		panic(err)
	}

	return coins
}

func (c *Context) IntervalSessions() time.Duration {
	return time.Duration(c.Config().Node.IntervalSessions)
}

func (c *Context) IntervalStatus() time.Duration {
	return time.Duration(c.Config().Node.IntervalStatus)
}
