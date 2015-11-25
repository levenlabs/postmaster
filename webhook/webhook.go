package webhook

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/levenlabs/go-llog"

	"github.com/levenlabs/postmaster/config"
	"github.com/levenlabs/postmaster/db"
	"gopkg.in/validator.v2"
)

type WebhookEvent db.StatsJob

func init() {
	if config.WebhookAddr == "" {
		return
	}

	go func() {
		s := &http.Server{
			Addr:           config.WebhookAddr,
			Handler:        http.HandlerFunc(hookHandler),
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		llog.Info("listening for webhook", llog.KV{"addr": config.WebhookAddr})
		err := s.ListenAndServe()
		llog.Fatal("error listening for webhoook", llog.KV{"addr": config.WebhookAddr, "err": err})
	}()
}

func hookHandler(w http.ResponseWriter, r *http.Request) {
	kv := llog.KV{"ip": r.RemoteAddr}
	llog.Info("webhook request", kv)

	if r.Method != "POST" {
		kv["method"] = r.Method
		llog.Warn("webhook invalid http method", kv)
		http.Error(w, "Invalid HTTP Method", http.StatusMethodNotAllowed)
		return
	}

	if config.WebhookPassword != "" {
		_, password, authOk := r.BasicAuth()
		if !authOk || password != config.WebhookPassword {
			llog.Warn("webhook authorization failed", kv)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	decoder := json.NewDecoder(r.Body)
	var events []WebhookEvent
	err := decoder.Decode(&events)
	if err != nil || len(events) == 0 {
		kv["err"] = err
		llog.Warn("webhook failed to parse body", kv)
		http.Error(w, "Invalid POST Body", http.StatusBadRequest)
		return
	}

	for _, event := range events {
		kv["event"] = event
		llog.Info("webhook processing event", kv)

		if err := validator.Validate(event); err != nil {
			kv["err"] = err
			llog.Warn("webhook event failed validation", kv)
			return
		}

		contents, err := json.Marshal(event)
		if err != nil {
			kv["err"] = err
			llog.Error("webhook couldn't marshal event", kv)
			delete(kv, "err")
			continue
		}
		if err = db.StoreStatsJob(string(contents)); err != nil {
			kv["err"] = err
			llog.Error("webhook couldn't store stats job", kv)
			delete(kv, "err")
		}
	}
}
