package db

import (
	"fmt"
	. "testing"

	"github.com/levenlabs/golib/testutil"
	"github.com/levenlabs/postmaster/config"
	"github.com/mediocregopher/radix.v2/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	normalQueue = fmt.Sprintf("email-normal-%s", testutil.RandStr())
	statsQueue = fmt.Sprintf("stats-normal-%s", testutil.RandStr())
}

func TestStoreSendJob(t *T) {
	err := StoreSendJob("hello")
	require.Nil(t, err)

	r, err := redis.Dial("tcp", config.OKQAddr)
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
	err := StoreStatsJob("hello2")
	require.Nil(t, err)

	r, err := redis.Dial("tcp", config.OKQAddr)
	require.Nil(t, err)
	res, err := r.Cmd("QRPOP", statsQueue, "EX", 0).Array()
	require.Nil(t, err)

	//res should be [queueID, "hello"]
	assert.Equal(t, 2, len(res))
	cont, err := res[1].Str()
	require.Nil(t, err)
	assert.Equal(t, "hello2", cont)
}
