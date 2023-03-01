package lite

import (
	"io"
	"sync"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
)

type Client struct {
	*sync.Mutex
	tmlog.Logger
	client.Context
	tx.Factory
	remotes []string
}

func NewClient() *Client {
	return &Client{
		Mutex: &sync.Mutex{},
	}
}

func NewDefaultClient() *Client {
	var (
		cfg = EncodingConfig()
		ctx = client.Context{}.
			WithBroadcastMode(flags.BroadcastBlock).
			WithCodec(cfg.Marshaler).
			WithInterfaceRegistry(cfg.InterfaceRegistry).
			WithLegacyAmino(cfg.Amino).
			WithOutputFormat("text").
			WithOutput(io.Discard).
			WithSkipConfirmation(true)
	)

	return NewClient().
		WithContext(ctx).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithTxConfig(cfg.TxConfig)
}

func (c *Client) WithContext(v client.Context) *Client {
	c.Context = v
	return c
}

func (c *Client) WithLogger(v tmlog.Logger) *Client {
	c.Logger = v
	return c
}

func (c *Client) WithRemotes(v []string) *Client {
	c.remotes = v
	return c
}

func (c *Client) WithAccountRetriever(v client.AccountRetriever) *Client {
	c.Context = c.Context.WithAccountRetriever(v)
	c.Factory = c.Factory.WithAccountRetriever(v)
	return c
}

func (c *Client) WithTxConfig(v client.TxConfig) *Client {
	c.Context = c.Context.WithTxConfig(v)
	c.Factory = c.Factory.WithTxConfig(v)
	return c
}

func (c *Client) WithChainID(v string) *Client {
	c.Context = c.Context.WithChainID(v)
	c.Factory = c.Factory.WithChainID(v)
	return c
}

func (c *Client) WithFeeGranterAddress(v sdk.AccAddress) *Client {
	c.Context = c.Context.WithFeeGranterAddress(v)
	return c
}

func (c *Client) WithFromAddress(v sdk.AccAddress) *Client {
	c.Context = c.Context.WithFromAddress(v)
	return c
}

func (c *Client) WithFromName(v string) *Client {
	c.Context = c.Context.WithFromName(v)
	return c
}

func (c *Client) WithKeyring(v keyring.Keyring) *Client {
	c.Context = c.Context.WithKeyring(v)
	c.Factory = c.Factory.WithKeybase(v)
	return c
}

func (c *Client) WithGas(v uint64) *Client {
	c.Factory = c.Factory.WithGas(v)
	return c
}

func (c *Client) WithSimulateAndExecute(v bool) *Client {
	c.Factory = c.Factory.WithSimulateAndExecute(v)
	return c
}

func (c *Client) WithGasAdjustment(v float64) *Client {
	c.Factory = c.Factory.WithGasAdjustment(v)
	return c
}

func (c *Client) WithGasPrices(v string) *Client {
	c.Factory = c.Factory.WithGasPrices(v)
	return c
}

func (c *Client) WithSignModeStr(v string) *Client {
	m := signing.SignMode_SIGN_MODE_UNSPECIFIED
	switch v {
	case flags.SignModeDirect:
		m = signing.SignMode_SIGN_MODE_DIRECT
	case flags.SignModeLegacyAminoJSON:
		m = signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	}

	c.Context = c.Context.WithSignModeStr(v)
	c.Factory = c.Factory.WithSignMode(m)
	return c
}
