package webhook

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/levenlabs/go-llog"
	"github.com/levenlabs/golib/genapi"
	"github.com/levenlabs/postmaster/db"
	"github.com/levenlabs/postmaster/ga"
	"gopkg.in/validator.v2"
)

var webhookPassword string

// WebhookEvent is just a wrapper around db.StatsJob for now
// it holds a representation of an incoming webhook event
type WebhookEvent db.StatsJob

func init() {
	ga.GA.AppendInit(func(g *genapi.GenAPI) {
		addr, _ := g.ParamStr("--webhook-addr")
		if addr == "" {
			return
		}
		webhookPassword, _ = g.ParamStr("--webhook-pass")

		go func() {
			s := &http.Server{
				Addr:           addr,
				Handler:        http.HandlerFunc(hookHandler),
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				MaxHeaderBytes: 1 << 20,
			}
			llog.Info("listening for webhook", llog.KV{"addr": addr})
			err := s.ListenAndServe()
			llog.Fatal("error listening for webhoook", llog.KV{"addr": addr}, llog.ErrKV(err))
		}()
	})
}

func hookHandler(w http.ResponseWriter, r *http.Request) {
	kv := llog.KV{"ip": r.RemoteAddr}
	llog.Debug("webhook request", kv)

	if r.Method != "POST" {
		kv["method"] = r.Method
		llog.Warn("webhook invalid http method", kv)
		http.Error(w, "Invalid HTTP Method", http.StatusMethodNotAllowed)
		return
	}

	if webhookPassword != "" {
		_, password, authOk := r.BasicAuth()
		if !authOk || password != webhookPassword {
			llog.Warn("webhook authorization failed", kv)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	decoder := json.NewDecoder(r.Body)
	var events []WebhookEvent
	err := decoder.Decode(&events)
	if err != nil || len(events) == 0 {
		llog.Warn("webhook failed to parse body", kv, llog.ErrKV(err))
		http.Error(w, "Invalid POST Body", http.StatusBadRequest)
		return
	}

	for _, event := range events {
		kv["event"] = event
		llog.Debug("webhook processing event", kv)

		// assume everything pre-environment is from production
		if event.SentEnvironment == "" {
			event.SentEnvironment = "production"
		}
		if event.SentEnvironment != "production" && event.SentEnvironment != "staging" {
			llog.Info("dropping webhook from non-production and non-staging environment", kv.Set("env", event.SentEnvironment))
			continue
		}

		if event.StatsID == "" && event.OldStatsID != "" {
			event.StatsID = event.OldStatsID
		}
		if err := validator.Validate(event); err != nil {
			llog.Warn("webhook event failed validation", kv, llog.ErrKV(err))
			continue
		}

		contents, err := json.Marshal(event)
		if err != nil {
			llog.Error("webhook couldn't marshal event", kv, llog.ErrKV(err))
			continue
		}
		if err = db.StoreStatsJob(string(contents)); err != nil {
			llog.Error("webhook couldn't store stats job", kv, llog.ErrKV(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
