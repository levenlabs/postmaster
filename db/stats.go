package db

import (
	"time"

	"github.com/levenlabs/go-llog"
	"github.com/levenlabs/golib/rpcutil"
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

type StatsJob struct {
	Email     string `json:"email" validate:"email,nonzero"` //Email address of the intended recipient
	Timestamp int64  `json:"timestamp,omitempty"`
	Type      string `json:"event" validate:"nonzero"`    //One of: bounce, deferred, delivered, dropped, processed
	StatsID   string `json:"stats_id" validate:"nonzero"` //json key must match db.UniqueArgID
	Reason    string `json:"reason,omitempty" validate:"max=1024"`
}

type StatDoc struct {
	ID         bson.ObjectId `bson:"_id,omitempty"`
	Recipient  string        `bson:"r"`
	EmailFlags int64         `bson:"ef"`
	StateFlags int64         `bson:"s"`
	TSCreated  time.Time     `bson:"tc"`
	TSUpdated  time.Time     `bson:"ts"`
	Error      string        `bson:"err,omitempty"`
}

func init() {
	rpcutil.InstallCustomValidators()
}

func GenerateEmailID(recipient string, flags int64) string {
	if mongoDisabled {
		return ""
	}
	now := time.Now()
	doc := &StatDoc{Recipient: recipient, EmailFlags: flags, TSCreated: now, TSUpdated: now}
	//generate our own ObjectID since mgo doesn't do it for insert
	doc.ID = bson.NewObjectId()
	var err error
	statsSH.WithColl(func(c *mgo.Collection) {
		err = c.Insert(doc)
	})
	if err != nil {
		llog.Error("error inserting in GenerateEmailID", llog.KV{"doc": doc, "err": err})
		return ""
	}
	return doc.ID.Hex()
}

//this is mostly for testing purposes
func GetStats(id string) *StatDoc {
	if mongoDisabled {
		return nil
	}
	doc := &StatDoc{}
	statsSH.WithColl(func(c *mgo.Collection) {
		c.FindId(bson.ObjectIdHex(id)).One(doc)
	})
	return doc
}

func markAs(id string, flag int, reason string) {
	if mongoDisabled {
		return
	}
	if !bson.IsObjectIdHex(id) {
		llog.Warn("invalid id sent to markAs", llog.KV{"id": id, "flag": flag, "reason": reason})
		return
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
	if err != nil {
		llog.Warn("error updating in markAs", llog.KV{"err": err, "id": id, "flag": flag, "reason": reason})
	}
}

func MarkAsDelivered(id string) {
	markAs(id, Delivered, "")
}

func MarkAsBounced(id string, reason string) {
	markAs(id, Bounced, reason)
}

func MarkAsDropped(id string, reason string) {
	markAs(id, Dropped, reason)
}

func MarkAsSpamReported(id string) {
	markAs(id, SpamReported, "")
}

func MarkAsOpened(id string) {
	markAs(id, Opened, "")
}
