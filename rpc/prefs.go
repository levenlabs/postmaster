package rpc

import (
	"github.com/levenlabs/golib/rpcutil"
	"github.com/levenlabs/postmaster/db"
	"net/http"
)

func init() {
	rpcutil.InstallCustomValidators()
}

type UpdatePrefsArgs struct {
	Email string `json:"email" validate:"email,nonzero"`
	Flags int64  `json:"flags" validate:"nonzero"`
}

type UpdatePrefsResult struct {
	Success bool `json:"success"`
}

func (_ Postmaster) UpdatePrefs(r *http.Request, args *UpdatePrefsArgs, reply *UpdatePrefsResult) error {
	if err := db.StoreEmailFlags(args.Email, args.Flags); err != nil {
		return err
	}
	reply.Success = true
	return nil
}
