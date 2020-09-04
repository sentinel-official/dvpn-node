package types

import (
	"context"

	"github.com/tendermint/tendermint/libs/log"
)

var (
	ctx = NewContext()
)

func GetContext() *Context {
	return ctx
}

type Context struct {
	ctx    context.Context
	logger log.Logger
	config *Config
	sealed bool
}

func NewContext() *Context {
	return &Context{
		ctx:    context.Background(),
		config: NewConfig(),
	}
}

func (c *Context) assert() {
	if c.sealed {
		panic("context is sealed")
	}
}

func (c *Context) WithContext(ctx context.Context) *Context {
	c.assert()

	c.ctx = ctx
	return c
}

func (c *Context) WithLogger(logger log.Logger) *Context {
	c.assert()

	c.logger = logger
	return c
}

func (c *Context) WithConfig(config *Config) *Context {
	c.assert()

	c.config = config
	return c
}

func (c *Context) WithValue(key, value interface{}) *Context {
	c.assert()

	c.WithContext(context.WithValue(c.ctx, key, value))
	return c
}

func (c *Context) Seal() {
	c.sealed = true
}

func (c *Context) Context() context.Context          { return c.ctx }
func (c *Context) Logger() log.Logger                { return c.logger }
func (c *Context) Config() *Config                   { return c.config }
func (c *Context) Sealed() bool                      { return c.sealed }
func (c *Context) Value(key interface{}) interface{} { return c.ctx.Value(key) }
