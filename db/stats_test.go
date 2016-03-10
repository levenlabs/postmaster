package db

import (
	"github.com/levenlabs/golib/testutil"
	"github.com/levenlabs/golib/timeutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/validator.v2"
	. "testing"
	"time"
)

func init() {
	RandomizeColls()
}

func TestGenerateEmailID(t *T) {
	if mongoDisabled {
		return
	}
	id := GenerateEmailID("test@test", 1, "", "production")
	require.NotEmpty(t, id)

	doc, err := GetStats(id)
	require.Nil(t, err)

	assert.NotNil(t, doc.ID)
	assert.Equal(t, id, doc.ID.Hex())
	assert.Equal(t, "test@test", doc.Recipient)
	assert.Equal(t, int64(1), doc.EmailFlags)
}

func TestDeleteEmailID(t *T) {
	if mongoDisabled {
		return
	}
	id := GenerateEmailID("test@test", 1, "", "production")
	require.NotEmpty(t, id)

	err := removeEmailID(id)
	require.Nil(t, err)

	_, err = GetStats(id)
	assert.NotNil(t, err)
}

func TestMarkAs(t *T) {
	if mongoDisabled {
		return
	}
	id := GenerateEmailID("test@test", 0, "", "production")
	require.NotEmpty(t, id)

	markAs(id, 9, "Test")

	doc, err := GetStats(id)
	require.Nil(t, err)

	assert.Equal(t, int64(9), doc.StateFlags)
	assert.Equal(t, "Test", doc.Error)
}

func TestValidation(t *T) {
	if mongoDisabled {
		return
	}
	s := &StatsJob{}
	assert.NotNil(t, validator.Validate(s))

	s = &StatsJob{
		Email:           "fake",
		Timestamp:       timeutil.TimestampNow(),
		Type:            testutil.RandStr(),
		StatsID:         testutil.RandStr(),
		SentEnvironment: "production",
	}
	assert.NotNil(t, validator.Validate(s))

	s = &StatsJob{
		Email:           "fake@test",
		Timestamp:       timeutil.TimestampNow(),
		Type:            testutil.RandStr(),
		StatsID:         testutil.RandStr(),
		SentEnvironment: "production",
	}
	assert.Nil(t, validator.Validate(s))

	s = &StatsJob{
		Email:           "fake@test",
		Timestamp:       timeutil.TimestampNow(),
		StatsID:         testutil.RandStr(),
		SentEnvironment: "production",
	}
	assert.NotNil(t, validator.Validate(s))

	s = &StatsJob{
		Email:           "fake@test",
		Timestamp:       timeutil.TimestampNow(),
		Type:            testutil.RandStr(),
		SentEnvironment: "production",
	}
	assert.NotNil(t, validator.Validate(s))

	s = &StatsJob{
		Email:     "fake@test",
		Timestamp: timeutil.TimestampNow(),
		StatsID:   testutil.RandStr(),
	}
	assert.NotNil(t, validator.Validate(s))
}

func TestGetLastUniqueID(t *T) {
	if mongoDisabled {
		return
	}
	email := "test@test"
	id := GenerateEmailID(email, 1, "hey", "production")
	require.NotEmpty(t, id)

	doc, err := GetLastUniqueID(email, "hey")
	require.Nil(t, err)
	assert.Equal(t, id, doc.ID.Hex())

	doc, err = GetLastUniqueID(email, "blah")
	require.NotNil(t, err)
	assert.Nil(t, doc)

	time.Sleep(2 * time.Second)

	id = GenerateEmailID(email, 1, "hey", "production")
	require.NotEmpty(t, id)

	doc, err = GetLastUniqueID(email, "hey")
	require.Nil(t, err)
	assert.Equal(t, id, doc.ID.Hex())
}
