package db

import (
	"github.com/levenlabs/golib/testutil"
	"github.com/stretchr/testify/assert"
	"gopkg.in/validator.v2"
	. "testing"
)

func init() {
	RandomizeColls()
}

func TestGenerateEmailID(t *T) {
	id := GenerateEmailID("test@test.com", 1)
	doc := GetStats(id)

	assert.NotNil(t, doc.ID)
	assert.Equal(t, id, doc.ID.Hex())
	assert.Equal(t, "test@test.com", doc.Recipient)
	assert.Equal(t, int64(1), doc.EmailFlags)
}

func TestMarkAs(t *T) {
	id := GenerateEmailID("test@test.com", 0)
	markAs(id, 9, "Test")
	doc := GetStats(id)

	assert.Equal(t, int64(9), doc.StateFlags)
	assert.Equal(t, "Test", doc.Error)
}

func TestValidation(t *T) {
	s := &StatsJob{}
	assert.NotNil(t, validator.Validate(s))

	s = &StatsJob{
		Email:     "fake",
		Timestamp: testutil.RandInt64(),
		Type:      testutil.RandStr(),
		StatsID:   testutil.RandStr(),
	}
	assert.NotNil(t, validator.Validate(s))

	s = &StatsJob{
		Email:     "fake@gmail.com",
		Timestamp: testutil.RandInt64(),
		Type:      testutil.RandStr(),
		StatsID:   testutil.RandStr(),
	}
	assert.Nil(t, validator.Validate(s))

	s = &StatsJob{
		Email:     "fake@gmail.com",
		Timestamp: testutil.RandInt64(),
		StatsID:   testutil.RandStr(),
	}
	assert.NotNil(t, validator.Validate(s))

	s = &StatsJob{
		Email:     "fake@gmail.com",
		Timestamp: testutil.RandInt64(),
		Type:      testutil.RandStr(),
	}
	assert.NotNil(t, validator.Validate(s))
}
