package db

import (
	"fmt"
	. "testing"
	"time"

	"github.com/levenlabs/golib/testutil"
	"github.com/levenlabs/postmaster/ga"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var okqAddr string

func init() {
	useOkq = true
	okqAddr, _ = ga.GA.ParamStr("--okq-addr")
}

func TestStoreSendJob(t *T) {
	require.True(t, useOkq)
	// randomize the queue so the consumer that's consuming jobs doesn't pick
	// up our job and try to process it
	existingQueue := normalQueue
	defer func() {
		normalQueue = existingQueue
	}()
	normalQueue = fmt.Sprintf("%s", testutil.RandStr())

	err := StoreSendJob("hello")
	require.Nil(t, err)

	r, err := redis.DialTimeout("tcp", okqAddr, 5*time.Second)
	require.Nil(t, err)
	res, err := r.Cmd("QRPOP", normalQueue, "EX", 0).Array()
	require.Nil(t, err)

	//res should be [queueID, "hello"]
	assert.Equal(t, 2, len(res))
	cont, err := res[1].Str()
	require.Nil(t, err)
	assert.Equal(t, "hello", cont)
}

func TestStoreStatsJob(t *T) {
	require.True(t, useOkq)
	// randomize the queue so the consumer that's consuming jobs doesn't pick
	// up our job and try to process it
	existingQueue := statsQueue
	defer func() {
		statsQueue = existingQueue
	}()
	statsQueue = fmt.Sprintf("%s", testutil.RandStr())

	err := StoreStatsJob("hello2")
	require.Nil(t, err)

	r, err := redis.DialTimeout("tcp", okqAddr, 5*time.Second)
	require.Nil(t, err)
	res, err := r.Cmd("QRPOP", statsQueue, "EX", 0).Array()
	require.Nil(t, err)

	//res should be [queueID, "hello"]
	assert.Equal(t, 2, len(res))
	cont, err := res[1].Str()
	require.Nil(t, err)
	assert.Equal(t, "hello2", cont)
}
