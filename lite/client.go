package lite

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	crypto "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/lite"
	"github.com/tendermint/tendermint/lite/proxy"
	"github.com/tendermint/tendermint/rpc/client"

	"github.com/sentinel-official/dvpn-node/types"
)

type Client struct {
	ctx   context.CLIContext
	txb   auth.TxBuilder
	mutex sync.Mutex
}

func NewClient() *Client {
	return new(Client).
		WithOutput(os.Stdout).
		WithOutputFormat("text").
		WithHeight(0).
		WithUseLedger(false).
		WithBroadcastMode("block").
		WithSimulate(false).
		WithGenerateOnly(false).
		WithIndent(false).
		WithSkipConfirm(true).
		WithSimulateAndExecute(true).
		WithMemo("")
}

func NewClientFromConfig(cfg *types.Config) (*Client, error) {
	var (
		verifier lite.Verifier
		home     = viper.GetString(types.FlagHome)
		node     = client.NewHTTP(cfg.Chain.RPCAddress, "/websocket")
	)

	kb, err := keys.NewKeyBaseFromDir(home)
	if err != nil {
		return nil, err
	}

	info, err := kb.Get(cfg.Node.From)
	if err != nil {
		return nil, err
	}

	if !cfg.Chain.TrustNode {
		verifier, err = proxy.NewVerifier(cfg.Chain.ID, filepath.Join(home, "lite"), node, log.NewNopLogger(), 16)
		if err != nil {
			return nil, err
		}
	}

	return NewClient().
		WithGasAdjustment(cfg.Chain.GasAdjustment).
		WithNode(node).
		WithKeybase(kb).
		WithNodeURI(cfg.Chain.RPCAddress).
		WithFrom(cfg.Node.From).
		WithTrustNode(cfg.Chain.TrustNode).
		WithVerifier(verifier).
		WithVerifierHome(filepath.Join(home, "lite")).
		WithFromAddress(info.GetAddress()).
		WithFromName(cfg.Node.From).
		WithGas(cfg.Chain.Gas).
		WithChainID(cfg.Chain.ID).
		WithFees(cfg.Chain.Fees).
		WithGasPrices(cfg.Chain.GasPrices), nil
}

func (c *Client) WithCodec(v *codec.Codec) *Client         { c.ctx.Codec = v; return c }
func (c *Client) WithNode(v client.Client) *Client         { c.ctx.Client = v; return c }
func (c *Client) WithOutput(v io.Writer) *Client           { c.ctx.Output = v; return c }
func (c *Client) WithOutputFormat(s string) *Client        { c.ctx.OutputFormat = s; return c }
func (c *Client) WithHeight(s int64) *Client               { c.ctx.Height = s; return c }
func (c *Client) WithNodeURI(s string) *Client             { c.ctx.NodeURI = s; return c }
func (c *Client) WithFrom(s string) *Client                { c.ctx.From = s; return c }
func (c *Client) WithTrustNode(t bool) *Client             { c.ctx.TrustNode = t; return c }
func (c *Client) WithUseLedger(t bool) *Client             { c.ctx.UseLedger = t; return c }
func (c *Client) WithBroadcastMode(s string) *Client       { c.ctx.BroadcastMode = s; return c }
func (c *Client) WithVerifier(v lite.Verifier) *Client     { c.ctx.Verifier = v; return c }
func (c *Client) WithVerifierHome(s string) *Client        { c.ctx.VerifierHome = s; return c }
func (c *Client) WithSimulate(t bool) *Client              { c.ctx.Simulate = t; return c }
func (c *Client) WithGenerateOnly(t bool) *Client          { c.ctx.GenerateOnly = t; return c }
func (c *Client) WithFromName(s string) *Client            { c.ctx.FromName = s; return c }
func (c *Client) WithIndent(t bool) *Client                { c.ctx.Indent = t; return c }
func (c *Client) WithSkipConfirm(t bool) *Client           { c.ctx.SkipConfirm = t; return c }
func (c *Client) WithFromAddress(v sdk.AccAddress) *Client { c.ctx.FromAddress = v; return c }
func (c *Client) WithSequence(i uint64) *Client            { c.txb = c.txb.WithSequence(i); return c }
func (c *Client) WithGas(i uint64) *Client                 { c.txb = c.txb.WithGas(i); return c }
func (c *Client) WithChainID(s string) *Client             { c.txb = c.txb.WithChainID(s); return c }
func (c *Client) WithMemo(s string) *Client                { c.txb = c.txb.WithMemo(s); return c }
func (c *Client) WithFees(s string) *Client                { c.txb = c.txb.WithFees(s); return c }
func (c *Client) WithGasPrices(s string) *Client           { c.txb = c.txb.WithGasPrices(s); return c }
func (c *Client) WithTxEncoder(v sdk.TxEncoder) *Client    { c.txb = c.txb.WithTxEncoder(v); return c }
func (c *Client) WithAccountNumber(i uint64) *Client       { c.txb = c.txb.WithAccountNumber(i); return c }

func (c *Client) WithKeybase(v crypto.Keybase) *Client {
	c.ctx.Keybase = v
	c.txb = c.txb.WithKeybase(v)

	return c
}

func (c *Client) WithGasAdjustment(i float64) *Client {
	c.txb = auth.NewTxBuilder(
		c.txb.TxEncoder(),
		c.txb.AccountNumber(),
		c.txb.Sequence(),
		c.txb.Gas(),
		i,
		c.txb.SimulateAndExecute(),
		c.txb.ChainID(),
		c.txb.Memo(),
		c.txb.Fees(),
		c.txb.GasPrices(),
	)
	return c
}

func (c *Client) WithSimulateAndExecute(t bool) *Client {
	c.txb = auth.NewTxBuilder(
		c.txb.TxEncoder(),
		c.txb.AccountNumber(),
		c.txb.Sequence(),
		c.txb.Gas(),
		c.txb.GasAdjustment(),
		t,
		c.txb.ChainID(),
		c.txb.Memo(),
		c.txb.Fees(),
		c.txb.GasPrices(),
	)
	return c
}

func (c *Client) FromAddress() sdk.AccAddress { return c.ctx.FromAddress }
