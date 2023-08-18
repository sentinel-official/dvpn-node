package context

import (
	"net"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	hubtypes "github.com/sentinel-official/hub/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"gorm.io/gorm"

	geoiptypes "github.com/sentinel-official/dvpn-node/libs/geoip/types"
	"github.com/sentinel-official/dvpn-node/lite"
	"github.com/sentinel-official/dvpn-node/types"
)

type Context struct {
	bandwidth *hubtypes.Bandwidth
	client    *lite.Client
	config    *types.Config
	database  *gorm.DB
	handler   http.Handler
	location  *geoiptypes.GeoIPLocation
	logger    tmlog.Logger
	service   types.Service
}

func NewContext() *Context {
	return &Context{}
}

func (c *Context) WithBandwidth(v *hubtypes.Bandwidth) *Context      { c.bandwidth = v; return c }
func (c *Context) WithClient(v *lite.Client) *Context                { c.client = v; return c }
func (c *Context) WithConfig(v *types.Config) *Context               { c.config = v; return c }
func (c *Context) WithDatabase(v *gorm.DB) *Context                  { c.database = v; return c }
func (c *Context) WithHandler(v http.Handler) *Context               { c.handler = v; return c }
func (c *Context) WithLocation(v *geoiptypes.GeoIPLocation) *Context { c.location = v; return c }
func (c *Context) WithLogger(v tmlog.Logger) *Context                { c.logger = v; return c }
func (c *Context) WithService(v types.Service) *Context              { c.service = v; return c }

func (c *Context) Address() hubtypes.NodeAddress       { return c.Operator().Bytes() }
func (c *Context) Bandwidth() *hubtypes.Bandwidth      { return c.bandwidth }
func (c *Context) Client() *lite.Client                { return c.client }
func (c *Context) Config() *types.Config               { return c.config }
func (c *Context) Database() *gorm.DB                  { return c.database }
func (c *Context) Handler() http.Handler               { return c.handler }
func (c *Context) IntervalSetSessions() time.Duration  { return c.Config().Node.IntervalSetSessions }
func (c *Context) IntervalUpdateStatus() time.Duration { return c.Config().Node.IntervalUpdateStatus }
func (c *Context) ListenOn() string                    { return c.Config().Node.ListenOn }
func (c *Context) Location() *geoiptypes.GeoIPLocation { return c.location }
func (c *Context) Log() tmlog.Logger                   { return c.logger }
func (c *Context) Moniker() string                     { return c.Config().Node.Moniker }
func (c *Context) Operator() sdk.AccAddress            { return c.client.FromAddress() }
func (c *Context) RemoteURL() string                   { return c.Config().Node.RemoteURL }
func (c *Context) Service() types.Service              { return c.service }

func (c *Context) IntervalUpdateSessions() time.Duration {
	return c.Config().Node.IntervalUpdateSessions
}

func (c *Context) IPv4Address() net.IP {
	addr := c.Config().Node.IPv4Address
	if addr == "" {
		addr = c.Location().IP
	}

	return net.ParseIP(addr).To4()
}

func (c *Context) GigabytePrices() sdk.Coins {
	if c.Config().Node.GigabytePrices == "" {
		return nil
	}

	coins, err := sdk.ParseCoinsNormalized(c.Config().Node.GigabytePrices)
	if err != nil {
		panic(err)
	}

	return coins
}

func (c *Context) HourlyPrices() sdk.Coins {
	if c.Config().Node.HourlyPrices == "" {
		return nil
	}

	coins, err := sdk.ParseCoinsNormalized(c.Config().Node.HourlyPrices)
	if err != nil {
		panic(err)
	}

	return coins
}
