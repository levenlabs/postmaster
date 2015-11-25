// Package rpc contains all the methods exposed to the rpc interface
package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/levenlabs/go-llog"
	"github.com/levenlabs/golib/rpcutil"
	"github.com/levenlabs/postmaster/db"
	"github.com/levenlabs/postmaster/sender"
)

// EnqueueArgs defines the arguments for the Postmaster.Enqueue endpoint
type EnqueueArgs sender.Mail

// EnqueueResult defines the response object for the Postmaster.Enqueue endpoint
type EnqueueResult struct {
	Success bool `json:"success"`
}

// Create is an rpc method which creates a Campaign and persists it to disk
func (_ Postmaster) Enqueue(r *http.Request, args *EnqueueArgs, reply *EnqueueResult) error {
	kv := rpcutil.RequestKV(r)
	// validation of email addresses is done with the validation library
	// more advanced validation is done in validateEnqueueArgs
	if err := validateEnqueueArgs(args); err != nil {
		kv["err"] = err
		llog.Warn("badly formed Enqueue request", kv)
		return err
	}

	allowed := db.VerifyEmailAllowed(args.To, args.Flags)
	if !allowed {
		kv["flags"] = fmt.Sprintf("%b", args.Flags)
		llog.Warn("cannot send email due to flags", kv)
		//even though we didn't send it, it didn't fail, the user just doesn't want this email
		reply.Success = true
		return nil
	}

	contents, err := json.Marshal(args)
	if err != nil {
		return err
	}
	err = db.StoreSendJob(string(contents))
	if err != nil {
		return err
	}
	reply.Success = true
	return nil
}

func validateEnqueueArgs(args *EnqueueArgs) error {
	if strings.HasSuffix(args.To, "@test") {
		return errors.New("to address cannot end in @test.com")
	}
	if args.HTML == "" && args.Text == "" {
		return errors.New("you must send either html or text")
	}
	return nil
}
