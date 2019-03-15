package resource_callback

import "github.com/ecletus/admin"

// Chain returns a Callbacks type from a slice of middleware handlers.
func Chain(callbacks ...*Callback) Callbacks {
	return Callbacks(callbacks)
}

// Handler builds and returns a http.Handler from the chain of middlewares,
// with `h http.Handler` as the final handler.
func (mws Callbacks) Run(res *admin.Resource) {
	chain := &ChainHandler{Callbacks:mws, Resource:res}
	chain.Run()
}

// ChainHandler is a http.Handler with support for handler composition and
// execution.
type ChainHandler struct {
	Callbacks Callbacks
	Resource  *admin.Resource
	index int
}


func (c *ChainHandler) Run() {
	for i, cb := range c.Callbacks {
		c.index = i
		cb.Handler(c)
	}
}

func (c *ChainHandler) Callback() *Callback {
	return c.Callbacks[c.index]
}

func (c *ChainHandler) Index() int {
	return c.index
}