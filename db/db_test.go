package db

import (
	"fmt"

	"github.com/levenlabs/golib/testutil"
	"github.com/levenlabs/postmaster/ga"
)

// this file just puts GA in TestMode()
// it also randomizes the collections

func init() {
	emailsColl = fmt.Sprintf("emails-%s", testutil.RandStr())
	emailSH.Coll = emailsColl
	statsColl = fmt.Sprintf("records-%s", testutil.RandStr())
	statsSH.Coll = statsColl
	ga.GA.TestMode()
}
