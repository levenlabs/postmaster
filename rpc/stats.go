package rpc

import (
	"github.com/levenlabs/postmaster/db"
	"gopkg.in/mgo.v2"
	"net/http"
)

type GetLastEmailArgs struct {
	//this should match the validate for sender.Mail
	To       string `json:"to" validate:"email,nonzero,max=256"`
	UniqueID string `json:"uniqueID" validate:"nonzero,max=256"`
}

type GetLastEmailResult struct {
	Stat *db.StatDoc `json:"stat"`
}

// GetLastEmail gets stats for the last email sent for a specific unique ID
// If no records were found, {"stat": null} is returned
func (_ Postmaster) GetLastEmail(r *http.Request, args *GetLastEmailArgs, reply *GetLastEmailResult) error {
	doc, err := db.GetLastUniqueID(args.To, args.UniqueID)
	reply.Stat = doc
	// If it was a not found error then ignore that
	if err == mgo.ErrNotFound {
		return nil
	}
	return err
}
