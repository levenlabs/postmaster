// Package rpc contains all the methods exposed to the rpc interface
package rpc

import (
	"github.com/levenlabs/golib/rpcutil"
	"net/http"
)

type Postmaster struct{}

func init() {
	rpcutil.InstallCustomValidators()
}

// RPC returns an http.Handler which will handle the RPC requests
func RPC() http.Handler {
	c := rpcutil.NewLLCodec()
	c.ValidateInput = true
	return rpcutil.JSONRPC2Handler(c, Postmaster{})
}

type successResult struct {
	Success bool `json:"success"`
}
