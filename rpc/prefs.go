package rpc

import (
	"github.com/levenlabs/postmaster/db"
	"net/http"
)

type updatePrefsArgs struct {
	Email string `json:"email" validate:"email,nonzero"`
	Flags int64  `json:"flags" validate:"nonzero"`
}

// UpdatePrefs updates an email addresses email preferences
func (_ Postmaster) UpdatePrefs(r *http.Request, args *updatePrefsArgs, reply *successResult) error {
	if err := db.StoreEmailFlags(args.Email, args.Flags); err != nil {
		return err
	}
	reply.Success = true
	return nil
}
