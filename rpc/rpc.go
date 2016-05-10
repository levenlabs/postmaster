// Package rpc contains all the methods exposed to the rpc interface
package rpc

import (
	"github.com/levenlabs/postmaster/ga"
)

// Postmaster is just a holder for the RPC methods
type Postmaster struct{}

func init() {
	ga.GA.Services = append(ga.GA.Services, Postmaster{})
}

// SuccessResult holds just a Success bool and is used for methods that don't
// need to return anything
type SuccessResult struct {
	Success bool `json:"success"`
}
