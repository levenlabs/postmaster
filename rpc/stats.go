package rpc

import (
	"github.com/levenlabs/postmaster/db"
	"gopkg.in/mgo.v2"
	"net/http"
)

type getLastEmailArgs struct {
	Email    string `json:"email" validate:"email,nonzero"`
	UniqueID string `json:"uniqueID" validate:"nonzero"`
}

type getLastEmailResult struct {
	Stat *db.StatDoc `json:"stat"`
}

// GetLastEmail gets stats for the last email sent for a specific unique ID
// If no records were found, {"stat": null} is returned
func (_ Postmaster) GetLastEmail(r *http.Request, args *getLastEmailArgs, reply *getLastEmailResult) error {
	doc, err := db.GetLastUniqueID(args.Email, args.UniqueID)
	reply.Stat = doc
	// If it was a not found error then ignore that
	if err == mgo.ErrNotFound {
		return nil
	}
	return err
}
