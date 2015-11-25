package main

import (
	"net/http"

	"github.com/levenlabs/go-llog"
	"github.com/levenlabs/go-srvclient"
	"github.com/levenlabs/postmaster/config"
	"github.com/levenlabs/postmaster/rpc"
	_ "github.com/levenlabs/postmaster/webhook"
	"github.com/mediocregopher/skyapi/client"
)

func main() {
	llog.SetLevelFromString(config.LogLevel)
	llog.Info("starting postmaster")

	if config.SkyAPIAddr != "" {
		skyapiAddr, err := srvclient.SRV(config.SkyAPIAddr)
		if err != nil {
			llog.Fatal("srv lookup of skyapi failed", llog.KV{"err": err})
		}

		kv := llog.KV{"skyapiAddr": skyapiAddr}
		llog.Info("connecting to skyapi", kv)

		go func() {
			kv["err"] = client.ProvideOpts(client.Opts{
				SkyAPIAddr:        skyapiAddr,
				Service:           "postmaster",
				ThisAddr:          config.InternalAPIAddr,
				ReconnectAttempts: 3,
			})
			llog.Fatal("skyapi giving up reconnecting", kv)
		}()
	}

	llog.Info("listening on internal api addr", llog.KV{"addr": config.InternalAPIAddr})
	err := http.ListenAndServe(config.InternalAPIAddr, rpc.RPC())
	llog.Fatal("internal api listen failed", llog.KV{"addr": config.InternalAPIAddr, "err": err})
}
