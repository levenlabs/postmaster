package db

import (
	"errors"
	"time"

	"github.com/levenlabs/go-llog"
	"github.com/levenlabs/golib/genapi"
	"github.com/levenlabs/golib/mgoutil"
	"github.com/levenlabs/postmaster/ga"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// EmailDoc represents a doc of the email's preferences, bounces, spams
type EmailDoc struct {
	Email       string      `bson:"_id"`
	UnsubFlags  int64       `bson:"f"`
	Bounces     []time.Time `bson:"b"` //also includes *some* drops
	SpamReports []time.Time `bson:"s"`
	TSUpdated   time.Time   `bson:"ts"`
}

var (
	mongoDisabled bool
	emailSH       mgoutil.SessionHelper
	statsSH       mgoutil.SessionHelper
	emailsColl    = "emails"
	// its called records because stats is a reserved collection in mongo
	statsColl = "records"

	//MongoDisabledErr is returned when we need mongo but don't have it
	MongoDisabledErr = errors.New("mongo disabled")
)

func init() {
	ga.GA.AppendInit(func(g *genapi.GenAPI) {
		emailSH = g.MongoInfo.CollSH(emailsColl)
		if emailSH.Session == nil {
			mongoDisabled = true
			return
		}
		statsSH = g.MongoInfo.CollSH(statsColl)
		statsSH.MustEnsureIndexes(
			mgo.Index{Key: []string{"uid", "r", "tc"}, Sparse: true},
		)
	})
}

// VerifyEmailAllowed verifies that we're allowed to send an email with flags to
// recipient
func VerifyEmailAllowed(email string, flags int64) bool {
	if mongoDisabled {
		//if they didn't run with mongo then they must want to approve all emails
		return true
	}
	res := &EmailDoc{}
	var err error
	emailSH.WithColl(func(c *mgo.Collection) {
		err = c.FindId(email).One(res)
	})
	if err != nil {
		//if the error is a not found error then its allowed since its not explicitly blocked
		if err == mgo.ErrNotFound {
			return true
		}
		llog.Error("error searching for doc by email", llog.KV{"email": email, "err": err})
		return false
	}
	//if none of the flags are present then its allowed
	//we check == 0 (and not != flags) since we want to know if they blocked ANY of the flags
	return res.UnsubFlags&flags == 0
}

// StoreEmailFlags updates the email with new flags restrictions
func StoreEmailFlags(email string, flags int64) error {
	if mongoDisabled {
		return MongoDisabledErr
	}
	update := bson.M{"$set": bson.M{"f": flags, "ts": time.Now()}}
	var err error
	emailSH.WithColl(func(c *mgo.Collection) {
		_, err = c.UpsertId(email, update)
	})
	return err
}

// StoreEmailBounce stores a new time when the email bounced
func StoreEmailBounce(email string) error {
	if mongoDisabled {
		return MongoDisabledErr
	}
	n := time.Now()
	update := bson.M{"$push": bson.M{"b": n, "ts": n}}
	var err error
	emailSH.WithColl(func(c *mgo.Collection) {
		_, err = c.UpsertId(email, update)
	})
	return err
}

// StoreEmailSpam stores a new time when the email was spammed
func StoreEmailSpam(email string) error {
	if mongoDisabled {
		return MongoDisabledErr
	}
	n := time.Now()
	update := bson.M{"$push": bson.M{"s": n, "ts": n}}
	var err error
	emailSH.WithColl(func(c *mgo.Collection) {
		_, err = c.UpsertId(email, update)
	})
	return err
}

// MoveEmailPrefs and email's prefs to a new address
func MoveEmailPrefs(oldEmail, newEmail string) error {
	if mongoDisabled {
		return MongoDisabledErr
	}
	n := time.Now()
	var err error
	emailSH.WithColl(func(c *mgo.Collection) {
		var doc, doc2 EmailDoc
		if err = c.FindId(oldEmail).One(&doc); err != nil {
			// If err is notFound, nothing to move
			if err == mgo.ErrNotFound {
				err = nil
			}
			return
		}
		doc2.Email = newEmail
		doc2.TSUpdated = n
		doc2.UnsubFlags = doc.UnsubFlags
		_, err = c.UpsertId(newEmail, doc2)
	})
	return err
}

// GetEmailFlags returns the unsub flags of an email address
func GetEmailFlags(email string) (int64, error) {
	if mongoDisabled {
		return 0, MongoDisabledErr
	}
	var err error
	flags := int64(1)
	emailSH.WithColl(func(c *mgo.Collection) {
		var doc EmailDoc
		if err = c.FindId(email).One(&doc); err != nil {
			if err == mgo.ErrNotFound {
				// If not found, they're not unsubbed to anything
				err = nil
			}
			return
		}
		flags = doc.UnsubFlags
	})
	return flags, err
}
