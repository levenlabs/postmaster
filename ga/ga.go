// Package ga is just a package to hold the GenAPI instance
package ga

import (
	"github.com/levenlabs/go-llog"
	"github.com/levenlabs/golib/genapi"
	"github.com/mediocregopher/lever"
)

var (
	// Environment represents the current running environment
	Environment string
)

// GA is an instance of the GenAPI for this rpc service
var GA = genapi.GenAPI{
	Name: "postmaster",
	OkqInfo: &genapi.OkqInfo{
		Optional: true,
	},
	MongoInfo: &genapi.MongoInfo{
		DBName:   "postmaster",
		Optional: true,
	},
	LeverParams: []lever.Param{
		{
			Name:        "--sendgrid-key",
			Description: "Sendgrid API Key",
		},
		{
			Name:        "--sendgrid-ip-pool",
			Description: "Specify a SendGrid IP pool for all emails to come from",
			Default:     "",
		},
		{
			Name:        "--webhook-addr",
			Description: "Address to listen for webhooks from sendgrid on",
			Default:     "127.0.0.1:8993",
		},
		{
			Name:        "--webhook-pass",
			Description: "Password (basic auth) to require for the webhook",
			Default:     "",
		},
		{
			Name:        "--environment",
			Description: "Running environment. Only prod and staging webhooks are processed.",
			Default:     "dev",
		},
	},
	Init: func(g *genapi.GenAPI) {
		Environment, _ = g.ParamStr("--environment")
		if Environment == "" {
			llog.Fatal("--environment is required")
		}
	},
}
