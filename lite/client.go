package lite

import (
	"io"
	"os"
	"sync"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	tmlog "github.com/tendermint/tendermint/libs/log"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

type Client struct {
	ctx    client.Context
	txf    tx.Factory
	logger tmlog.Logger
	mutex  sync.Mutex
}

func NewClient() *Client {
	return &Client{}
}

func NewDefaultClient() *Client {
	return NewClient().
		WithBroadcastMode("block").
		WithGenerateOnly(false).
		WithHeight(0).
		WithOutput(os.Stdout).
		WithOutputFormat("text").
		WithUseLedger(false).
		WithSimulate(false).
		WithSkipConfirm(true).
		WithMemo("")
}

func (c *Client) Copy() *Client {
	return &Client{
		ctx: c.ctx,
		txf: c.txf,
	}
}

func (c *Client) WithBroadcastMode(v string) *Client           { c.ctx.BroadcastMode = v; return c }
func (c *Client) WithClient(v rpcclient.Client) *Client        { c.ctx.Client = v; return c }
func (c *Client) WithFrom(v string) *Client                    { c.ctx.From = v; return c }
func (c *Client) WithFromAddress(v sdk.AccAddress) *Client     { c.ctx.FromAddress = v; return c }
func (c *Client) WithFromName(v string) *Client                { c.ctx.FromName = v; return c }
func (c *Client) WithGenerateOnly(v bool) *Client              { c.ctx.GenerateOnly = v; return c }
func (c *Client) WithHeight(v int64) *Client                   { c.ctx.Height = v; return c }
func (c *Client) WithHomeDir(v string) *Client                 { c.ctx.HomeDir = v; return c }
func (c *Client) WithInput(v io.Reader) *Client                { c.ctx.Input = v; return c }
func (c *Client) WithJSONCodec(v codec.JSONCodec) *Client      { c.ctx.JSONCodec = v; return c }
func (c *Client) WithKeyringDir(v string) *Client              { c.ctx.KeyringDir = v; return c }
func (c *Client) WithLegacyAmino(v *codec.LegacyAmino) *Client { c.ctx.LegacyAmino = v; return c }
func (c *Client) WithLogger(v tmlog.Logger) *Client            { c.logger = v; return c }
func (c *Client) WithNodeURI(v string) *Client                 { c.ctx.NodeURI = v; return c }
func (c *Client) WithOffline(v bool) *Client                   { c.ctx.Offline = v; return c }
func (c *Client) WithOutput(v io.Writer) *Client               { c.ctx.Output = v; return c }
func (c *Client) WithOutputFormat(v string) *Client            { c.ctx.OutputFormat = v; return c }
func (c *Client) WithSimulate(v bool) *Client                  { c.ctx.Simulate = v; return c }
func (c *Client) WithSkipConfirm(v bool) *Client               { c.ctx.SkipConfirm = v; return c }
func (c *Client) WithUseLedger(v bool) *Client                 { c.ctx.UseLedger = v; return c }

func (c *Client) WithAccountRetriever(v client.AccountRetriever) *Client {
	c.ctx.AccountRetriever = v
	c.txf = c.txf.WithAccountRetriever(v)

	return c
}

func (c *Client) WithInterfaceRegistry(v codectypes.InterfaceRegistry) *Client {
	c.ctx.InterfaceRegistry = v
	return c
}

func (c *Client) WithAccountNumber(v uint64) *Client  { c.txf = c.txf.WithAccountNumber(v); return c }
func (c *Client) WithFees(v string) *Client           { c.txf = c.txf.WithFees(v); return c }
func (c *Client) WithGas(v uint64) *Client            { c.txf = c.txf.WithGas(v); return c }
func (c *Client) WithGasAdjustment(v float64) *Client { c.txf = c.txf.WithGasAdjustment(v); return c }
func (c *Client) WithGasPrices(v string) *Client      { c.txf = c.txf.WithGasPrices(v); return c }
func (c *Client) WithMemo(v string) *Client           { c.txf = c.txf.WithMemo(v); return c }
func (c *Client) WithSequence(v uint64) *Client       { c.txf = c.txf.WithSequence(v); return c }
func (c *Client) WithTimeoutHeight(v uint64) *Client  { c.txf = c.txf.WithTimeoutHeight(v); return c }

func (c *Client) WithSimulateAndExecute(v bool) *Client {
	c.txf = c.txf.WithSimulateAndExecute(v)
	return c
}

func (c *Client) WithChainID(v string) *Client {
	c.ctx.ChainID = v
	c.txf = c.txf.WithChainID(v)

	return c
}

func (c *Client) WithKeyring(v keyring.Keyring) *Client {
	c.ctx.Keyring = v
	c.txf = c.txf.WithKeybase(v)

	return c
}

func (c *Client) WithSignMode(v string) *Client {
	c.ctx.SignModeStr = v

	var mode signing.SignMode
	switch v {
	case flags.SignModeDirect:
		mode = signing.SignMode_SIGN_MODE_DIRECT
	case flags.SignModeLegacyAminoJSON:
		mode = signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	default:
		mode = signing.SignMode_SIGN_MODE_UNSPECIFIED
	}

	c.txf.WithSignMode(mode)
	return c
}

func (c *Client) WithTxConfig(v client.TxConfig) *Client {
	c.ctx.TxConfig = v
	c.txf = c.txf.WithTxConfig(v)

	return c
}

func (c *Client) BroadcastMode() string       { return c.ctx.BroadcastMode }
func (c *Client) ChainID() string             { return c.ctx.ChainID }
func (c *Client) Client() rpcclient.Client    { return c.ctx.Client }
func (c *Client) From() string                { return c.ctx.From }
func (c *Client) FromAddress() sdk.AccAddress { return c.ctx.FromAddress }
func (c *Client) Keyring() keyring.Keyring    { return c.ctx.Keyring }
func (c *Client) Log() tmlog.Logger           { return c.logger }
func (c *Client) TxConfig() client.TxConfig   { return c.ctx.TxConfig }
func (c *Client) SimulateAndExecute() bool    { return c.txf.SimulateAndExecute() }
