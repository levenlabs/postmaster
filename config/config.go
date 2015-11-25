// Package config parses command-line/environment/config file arguments and puts
// together the configuration of this instance, which is made available to other
// packages.
package config

import "github.com/mediocregopher/lever"

// Configurable variables which are made available
var (
	InternalAPIAddr string
	SkyAPIAddr      string
	OKQAddr         string
	SendGridAPIKey  string
	MongoAddr       string
	WebhookAddr     string
	WebhookPassword string
	LogLevel        string
)

func init() {
	l := lever.New("postmaster", nil)
	l.Add(lever.Param{
		Name:        "--internal-addr",
		Description: "Address to listen on for the internal api",
		Default:     "127.0.0.1:9093",
	})
	l.Add(lever.Param{
		Name:        "--skyapi-addr",
		Description: "Hostname of skyapi, to be looked up via a SRV request. Unset means don't register with skyapi",
	})
	l.Add(lever.Param{
		Name:        "--okq-addr",
		Description: "Address okq is listening on",
		Default:     "127.0.0.1:4777",
	})
	l.Add(lever.Param{
		Name:        "--sendgrid-key",
		Description: "Sendgrid API Key",
		Default:     "",
	})
	l.Add(lever.Param{
		Name:        "--mongo-addr",
		Description: "Address mongo is listening on and database to use",
		Default:     "127.0.0.1:27017",
	})
	l.Add(lever.Param{
		Name:        "--webhook-addr",
		Description: "Address to listen for webhooks from sendgrid on",
		Default:     "127.0.0.1:8993",
	})
	l.Add(lever.Param{
		Name:        "--webhook-pass",
		Description: "Password (basic auth) to require for the webhook",
		Default:     "",
	})
	l.Add(lever.Param{
		Name:        "--log-level",
		Description: "Minimum log level to show, either debug, info, warn, error, or fatal",
		Default:     "info",
	})
	l.Parse()

	InternalAPIAddr, _ = l.ParamStr("--internal-addr")
	SkyAPIAddr, _ = l.ParamStr("--skyapi-addr")
	OKQAddr, _ = l.ParamStr("--okq-addr")
	SendGridAPIKey, _ = l.ParamStr("--sendgrid-key")
	MongoAddr, _ = l.ParamStr("--mongo-addr")
	WebhookAddr, _ = l.ParamStr("--webhook-addr")
	WebhookPassword, _ = l.ParamStr("--webhook-pass")
	LogLevel, _ = l.ParamStr("--log-level")
}
