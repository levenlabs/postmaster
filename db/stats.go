package db

import (
	"time"

	"errors"
	"fmt"
	"github.com/levenlabs/go-llog"
	"github.com/levenlabs/golib/rpcutil"
	"github.com/levenlabs/golib/timeutil"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	Sent int = 1 << iota
	Delivered
	SpamReported
	Bounced
	Dropped
	Opened
)

var MongoDisabledErr = errors.New("mongo disabled")

// A StatsJob encompasses a okq job in response to a webhook event and is used
// to update the StatDoc for a specific email identified by StatsID
type StatsJob struct {
	//Email address of the intended recipient
	Email string `json:"email" validate:"email,nonzero"`

	Timestamp timeutil.Timestamp `json:"timestamp,omitempty"`

	//Type is one of: bounce, deferred, delivered, dropped, processed
	Type string `json:"event" validate:"nonzero"`

	//json flag must match db.UniqueArgID in okq.go
	StatsID string `json:"pmStatsID" validate:"nonzero"`

	// this is the previous json key name before we changed it to pmStatusID
	OldStatsID string `json:"stats_id"`

	// Reason is miscellaneous data for why it bounced, dropped, etc
	Reason string `json:"reason,omitempty" validate:"max=1024"`
}

// A StatDoc represents an email that was sent
type StatDoc struct {
	// ID is a unique identifier for this doc not to be confused by the
	// user-supplied uniqueID field
	ID bson.ObjectId `json:"-" bson:"_id,omitempty"`

	// Recipient is the email address of the recipient
	Recipient string `json:"recipient" bson:"r"`

	// EmailFlags were the originally flags sent when sending the email
	EmailFlags int64 `json:"emailFlags" bson:"ef"`

	// StateFlags represent the current state of the email
	StateFlags int64 `json:"stateFlags" bson:"s"`

	// UniqueID was the original uniqueID sent to us in rpc.Enqueue
	UniqueID string `json:"uniqueID" bson:"uid"`

	// TSCreated is the time that the email was sent
	TSCreated timeutil.Timestamp `json:"tsCreated" bson:"tc"`

	// TSUpdated is the last time this doc was updated
	TSUpdated timeutil.Timestamp `json:"tsUpdated" bson:"ts"`

	// Error is the reason for why the email errored
	Error string `json:"error" bson:"err,omitempty"`
}

func init() {
	rpcutil.InstallCustomValidators()
}

// GenerateEmailID generates a uniqueID and stores a record of an intended email
// this is used in okq.go and in tests
func GenerateEmailID(recipient string, flags int64, uid string) string {
	if mongoDisabled {
		return ""
	}
	now := timeutil.TimestampNow()
	doc := &StatDoc{
		Recipient:  recipient,
		EmailFlags: flags,
		UniqueID:   uid,
		TSCreated:  now,
		TSUpdated:  now,
	}
	//generate our own ObjectID since mgo doesn't do it for insert
	doc.ID = bson.NewObjectId()
	var err error
	statsSH.WithColl(func(c *mgo.Collection) {
		err = c.Insert(doc)
	})
	if err != nil {
		llog.Error("error inserting in generateEmailID", llog.KV{"doc": doc, "err": err})
		return ""
	}
	return doc.ID.Hex()
}

// removeEmailID is used to remove an emailID if an email failed to
func removeEmailID(id string) error {
	var err error
	oid := bson.ObjectIdHex(id)
	statsSH.WithColl(func(c *mgo.Collection) {
		err = c.RemoveId(oid)
	})
	return err
}

//this is mostly for testing purposes
func GetStats(id string) (*StatDoc, error) {
	if mongoDisabled {
		return nil, MongoDisabledErr
	}
	var err error
	doc := &StatDoc{}
	statsSH.WithColl(func(c *mgo.Collection) {
		err = c.FindId(bson.ObjectIdHex(id)).One(doc)
	})

	if err != nil {
		doc = nil
	}
	return doc, err
}

func markAs(id string, flag int, reason string) error {
	if mongoDisabled {
		return MongoDisabledErr
	}
	if !bson.IsObjectIdHex(id) {
		llog.Warn("invalid id sent to markAs", llog.KV{"id": id, "flag": flag, "reason": reason})
		return fmt.Errorf("invalid id sent to markAs: %s", id)
	}

	var set map[string]interface{}
	if reason != "" {
		set = bson.M{"ts": time.Now(), "err": reason}
	} else {
		set = bson.M{"ts": time.Now()}
	}
	update := bson.M{"$bit": bson.M{"s": bson.M{"or": flag}}, "$set": set}
	var err error
	statsSH.WithColl(func(c *mgo.Collection) {
		err = c.UpdateId(bson.ObjectIdHex(id), update)
	})
	return err
}

func MarkAsDelivered(id string) error {
	return markAs(id, Delivered, "")
}

func MarkAsBounced(id string, reason string) error {
	return markAs(id, Bounced, reason)
}

func MarkAsDropped(id string, reason string) error {
	return markAs(id, Dropped, reason)
}

func MarkAsSpamReported(id string) error {
	return markAs(id, SpamReported, "")
}

func MarkAsOpened(id string) error {
	return markAs(id, Opened, "")
}

// GetLastUniqueID gets the last StatDoc for the given recipient and uniqueID
func GetLastUniqueID(recipient, uid string) (*StatDoc, error) {
	if mongoDisabled {
		return nil, MongoDisabledErr
	}
	var err error
	doc := &StatDoc{}
	statsSH.WithColl(func(c *mgo.Collection) {
		q := bson.M{
			"uid": uid,
			"r":   recipient,
		}
		// sort by the highest (newest) created times at top
		err = c.Find(q).Sort("-tc").One(doc)
	})
	if err != nil {
		doc = nil
	}
	return doc, err
}
