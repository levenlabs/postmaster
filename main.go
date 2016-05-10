package main

import (
	"github.com/levenlabs/postmaster/ga"
	_ "github.com/levenlabs/postmaster/rpc"
	_ "github.com/levenlabs/postmaster/webhook"
)

func main() {
	ga.GA.APIMode()
}
