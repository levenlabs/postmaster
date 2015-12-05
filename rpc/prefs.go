package rpc

import (
	"github.com/levenlabs/postmaster/db"
	"net/http"
)

type UpdatePrefsArgs struct {
	Email string `json:"email" validate:"email,nonzero"`
	Flags int64  `json:"flags" validate:"nonzero"`
}

// UpdatePrefs updates an email addresses email preferences
func (_ Postmaster) UpdatePrefs(r *http.Request, args *UpdatePrefsArgs, reply *SuccessResult) error {
	if err := db.StoreEmailFlags(args.Email, args.Flags); err != nil {
		return err
	}
	reply.Success = true
	return nil
}
