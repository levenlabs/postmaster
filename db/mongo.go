package db

import (
	"errors"
	"time"

	"fmt"
	"github.com/levenlabs/go-llog"
	"github.com/levenlabs/golib/mgoutil"
	"github.com/levenlabs/golib/testutil"
	"github.com/levenlabs/postmaster/config"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Names of databases and collections in mongo
const (
	DB         = "postmaster"
	EmailsColl = "emails"
	// stats is a reserved collection in mongo
	StatsColl = "records"
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
)

func init() {
	if config.MongoAddr == "" {
		mongoDisabled = true
		return
	}

	s := mgoutil.EnsureSession(config.MongoAddr)

	emailSH = mgoutil.SessionHelper{
		Session: s,
		DB:      DB,
		Coll:    EmailsColl,
	}

	statsSH = mgoutil.SessionHelper{
		Session: s,
		DB:      DB,
		Coll:    StatsColl,
	}

	statsSH.MustEnsureIndexes(
		mgo.Index{Key: []string{"uid", "r", "tc"}, Sparse: true},
	)
}

// this is ONLY exported so webhook_test can use it
// todo: we should find a way around exporting this
func RandomizeColls() {
	if !mongoDisabled {
		statsSH.DB = "test"
		statsSH.Coll = fmt.Sprintf("records-%s", testutil.RandStr())
	}

	if !mongoDisabled {
		emailSH.DB = "test"
		emailSH.Coll = fmt.Sprintf("emails-%s", testutil.RandStr())
	}
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
		return errors.New("--mongo-addr required")
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
		return errors.New("--mongo-addr required")
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
		return errors.New("--mongo-addr required")
	}
	n := time.Now()
	update := bson.M{"$push": bson.M{"s": n, "ts": n}}
	var err error
	emailSH.WithColl(func(c *mgo.Collection) {
		_, err = c.UpsertId(email, update)
	})
	return err
}
