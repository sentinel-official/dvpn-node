package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"cosmossdk.io/math"
	cosmossdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sentinel-official/sentinel-go-sdk/core"
	"github.com/sentinel-official/sentinel-go-sdk/libs/geoip"
	"github.com/sentinel-official/sentinel-go-sdk/libs/oracle"
	sentinelsdk "github.com/sentinel-official/sentinel-go-sdk/types"
	sentinelhub "github.com/sentinel-official/sentinelhub/v12/types"
	"github.com/sentinel-official/sentinelhub/v12/types/v1"
	"gorm.io/gorm"
)

// Context defines the application context, holding configurations and shared components.
type Context struct {
	accAddr        cosmossdk.AccAddress
	apiAddrs       []string
	apiListenAddr  string
	client         *core.Client
	database       *gorm.DB
	dlSpeed        math.Int
	geoIPClient    geoip.Client
	gigabytePrices v1.Prices
	handshakeDNS   bool
	homeDir        string
	hourlyPrices   v1.Prices
	input          io.Reader
	location       *geoip.Location
	maxPeers       uint
	moniker        string
	oracleClient   oracle.Client
	remoteAddrs    []string
	rpcAddrs       []string
	service        sentinelsdk.ServerService
	ulSpeed        math.Int

	sealed bool

	fm  sync.RWMutex
	txm sync.Mutex
}

// NewContext creates a new Context instance with default values.
func NewContext() *Context {
	return &Context{
		dlSpeed: math.ZeroInt(),
		ulSpeed: math.ZeroInt(),
	}
}

// Seal marks the context as sealed, preventing further modifications.
func (c *Context) Seal() *Context {
	c.sealed = true

	return c
}

// AccAddr returns the transaction sender address set in the context.
func (c *Context) AccAddr() cosmossdk.AccAddress {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.accAddr.Bytes()
}

// APIAddrs returns the api addresses set in the context.
func (c *Context) APIAddrs() []string {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.apiAddrs
}

// APIListenAddr returns the listen address of the node API.
func (c *Context) APIListenAddr() string {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.apiListenAddr
}

// Client returns the client instance set in the context.
func (c *Context) Client() *core.Client {
	c.fm.RLock()
	defer c.fm.RUnlock()

	c.client.SetRPCAddr(c.RPCAddr())

	return c.client
}

// Database returns the database connection set in the context.
func (c *Context) Database() *gorm.DB {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.database
}

// DatabaseFile returns the database path of the node.
func (c *Context) DatabaseFile() string {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return filepath.Join(c.HomeDir(), "data.db")
}

// GeoIPClient returns the GeoIP client set in the context.
func (c *Context) GeoIPClient() geoip.Client {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.geoIPClient
}

// GigabytePrices returns the gigabyte prices for nodes.
func (c *Context) GigabytePrices() v1.Prices {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.gigabytePrices
}

// HandshakeDNS reports whether the node advertises Handshake (HNS) DNS resolution.
func (c *Context) HandshakeDNS() bool {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.handshakeDNS
}

// HomeDir returns the home directory set in the context.
func (c *Context) HomeDir() string {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.homeDir
}

// HourlyPrices returns the hourly prices for nodes.
func (c *Context) HourlyPrices() v1.Prices {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.hourlyPrices
}

// Input returns the keyring input set in the context.
func (c *Context) Input() io.Reader {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.input
}

// Location returns the geolocation data set in the context.
func (c *Context) Location() *geoip.Location {
	c.fm.RLock()
	defer c.fm.RUnlock()

	if c.location == nil {
		return &geoip.Location{}
	}

	return c.location
}

// MaxPeers returns the maximum peers for the service.
func (c *Context) MaxPeers() uint {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.maxPeers
}

// Moniker returns the name or identifier for the node.
func (c *Context) Moniker() string {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.moniker
}

func (c *Context) NodeAddr() sentinelhub.NodeAddress {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.accAddr.Bytes()
}

// OracleClient returns the oracle client set in the context.
func (c *Context) OracleClient() oracle.Client {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.oracleClient
}

// RemoteAddrs returns the remote addresses set in the context.
func (c *Context) RemoteAddrs() []string {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.remoteAddrs
}

// RPCAddr returns the first RPC address from the list or an empty string if no addresses are available.
func (c *Context) RPCAddr() string {
	c.fm.RLock()
	defer c.fm.RUnlock()

	addrs := c.RPCAddrs()
	if len(addrs) == 0 {
		panic(errors.New("rpc_addrs is empty"))
	}

	return addrs[0]
}

// RPCAddrs returns the RPC addresses used for queries in the context.
func (c *Context) RPCAddrs() []string {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.rpcAddrs
}

// Service returns the server service instance set in the context.
func (c *Context) Service() sentinelsdk.ServerService {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.service
}

// SpeedtestResults returns the download and upload speeds set in the context.
func (c *Context) SpeedtestResults() (dlSpeed, ulSpeed math.Int) {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return c.dlSpeed, c.ulSpeed
}

// TLSCertFile returns the TLS certificate path of the node API server.
func (c *Context) TLSCertFile() string {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return filepath.Join(c.HomeDir(), "tls.crt")
}

// TLSKeyFile returns the TLS key path of the node API server.
func (c *Context) TLSKeyFile() string {
	c.fm.RLock()
	defer c.fm.RUnlock()

	return filepath.Join(c.HomeDir(), "tls.key")
}

// SanitizedGigabytePrices returns gigabyte prices filtered to include only valid denominations.
func (c *Context) SanitizedGigabytePrices(ctx context.Context) (v1.Prices, error) {
	params, err := c.Client().NodeParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting node params: %w", err)
	}

	prices := c.sanitizePrices(
		c.GigabytePrices(),
		params.GetMinGigabytePrices(),
	)

	return prices, nil
}

// SanitizedHourlyPrices returns hourly prices filtered to include only valid denominations.
func (c *Context) SanitizedHourlyPrices(ctx context.Context) (v1.Prices, error) {
	params, err := c.Client().NodeParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting node params: %w", err)
	}

	prices := c.sanitizePrices(
		c.HourlyPrices(),
		params.GetMinHourlyPrices(),
	)

	return prices, nil
}

// SetLocation sets the geolocation data in the context.
func (c *Context) SetLocation(location *geoip.Location) {
	c.fm.Lock()
	defer c.fm.Unlock()

	c.location = location
}

// SetRPCAddrs sets the RPC addresses in the context and allows for thread-safe updates.
func (c *Context) SetRPCAddrs(addrs []string) {
	c.fm.Lock()
	defer c.fm.Unlock()

	c.rpcAddrs = addrs
}

// SetSpeedtestResults sets the download and upload speeds in the context.
func (c *Context) SetSpeedtestResults(dlSpeed, ulSpeed math.Int) {
	c.fm.Lock()
	defer c.fm.Unlock()

	c.dlSpeed = dlSpeed
	c.ulSpeed = ulSpeed
}

// WithAccAddr sets the transaction sender address in the context and returns the updated context.
func (c *Context) WithAccAddr(addr cosmossdk.AccAddress) *Context {
	c.checkSealed()
	c.accAddr = addr

	return c
}

// WithAPIAddrs sets the api addresses in the context and returns the updated context.
func (c *Context) WithAPIAddrs(addrs []string) *Context {
	c.checkSealed()
	c.apiAddrs = addrs

	return c
}

// WithAPIListenAddr sets the listen address for the node API and returns the updated context.
func (c *Context) WithAPIListenAddr(addr string) *Context {
	c.checkSealed()
	c.apiListenAddr = addr

	return c
}

// WithClient sets the core client in the context and returns the updated context.
func (c *Context) WithClient(client *core.Client) *Context {
	c.checkSealed()
	c.client = client

	return c
}

// WithDatabase sets the database connection in the context and returns the updated context.
func (c *Context) WithDatabase(database *gorm.DB) *Context {
	c.checkSealed()
	c.database = database

	return c
}

// WithGeoIPClient sets the GeoIP client in the context and returns the updated context.
func (c *Context) WithGeoIPClient(client geoip.Client) *Context {
	c.checkSealed()
	c.geoIPClient = client

	return c
}

// WithGigabytePrices sets the gigabyte prices for nodes and returns the updated context.
func (c *Context) WithGigabytePrices(prices v1.Prices) *Context {
	c.checkSealed()
	c.gigabytePrices = prices

	return c
}

// WithHandshakeDNS sets the Handshake DNS availability flag and returns the updated context.
func (c *Context) WithHandshakeDNS(v bool) *Context {
	c.checkSealed()
	c.handshakeDNS = v

	return c
}

// WithHomeDir sets the home directory in the context and returns the updated context.
func (c *Context) WithHomeDir(dir string) *Context {
	c.checkSealed()
	c.homeDir = dir

	return c
}

// WithHourlyPrices sets the hourly prices for nodes and returns the updated context.
func (c *Context) WithHourlyPrices(prices v1.Prices) *Context {
	c.checkSealed()
	c.hourlyPrices = prices

	return c
}

// WithInput sets the keyring input in the context and returns the updated context.
func (c *Context) WithInput(input io.Reader) *Context {
	c.checkSealed()
	c.input = input

	return c
}

// WithMaxPeers sets maximum peers for the service and returns the updated context.
func (c *Context) WithMaxPeers(maxPeers uint) *Context {
	c.checkSealed()
	c.maxPeers = maxPeers

	return c
}

// WithMoniker sets the name or identifier for the node and returns the updated context.
func (c *Context) WithMoniker(moniker string) *Context {
	c.checkSealed()
	c.moniker = moniker

	return c
}

// WithOracleClient sets the oracle client in the context and returns the updated context.
func (c *Context) WithOracleClient(client oracle.Client) *Context {
	c.checkSealed()
	c.oracleClient = client

	return c
}

// WithRemoteAddrs sets the remote addresses in the context and returns the updated context.
func (c *Context) WithRemoteAddrs(addrs []string) *Context {
	c.checkSealed()
	c.remoteAddrs = addrs

	return c
}

// WithRPCAddrs sets the RPC addresses for queries in the context and returns the updated context.
func (c *Context) WithRPCAddrs(addrs []string) *Context {
	c.checkSealed()
	c.rpcAddrs = addrs

	return c
}

// WithService sets the server service in the context and returns the updated context.
func (c *Context) WithService(service sentinelsdk.ServerService) *Context {
	c.checkSealed()
	c.service = service

	return c
}

// checkSealed verifies if the context is sealed to prevent modification.
func (c *Context) checkSealed() {
	if c.sealed {
		panic(errors.New("context is sealed"))
	}
}

// sanitizePrices filters and validates a set of prices against the minimum required prices.
func (c *Context) sanitizePrices(prices v1.Prices, minPrices v1.Prices) (newPrices v1.Prices) {
	m := minPrices.Map()
	for _, price := range prices {
		if _, ok := m[price.Denom]; ok {
			newPrices = newPrices.Add(price)
		}
	}

	return
}
