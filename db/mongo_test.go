package db

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2"
	. "testing"
	"time"
)

func TestStoreEmailFlags(t *T) {
	require.False(t, mongoDisabled)
	emailSH.WithColl(func(c *mgo.Collection) {
		email := "test@test.com"
		err := StoreEmailFlags(email, 1)
		require.Nil(t, err)
		doc := &EmailDoc{}
		err = c.FindId(email).One(doc)
		require.Nil(t, err)

		assert.Equal(t, email, doc.Email)
		assert.Equal(t, int64(1), doc.UnsubFlags)
	})
}

func TestVerifyEmailAllowed(t *T) {
	require.False(t, mongoDisabled)
	email := "test1@test.com"
	err := StoreEmailFlags(email, 1)
	require.Nil(t, err)
	allowed := VerifyEmailAllowed(email, 1)
	assert.False(t, allowed)
	allowed = VerifyEmailAllowed(email, 2)
	assert.True(t, allowed)
	allowed = VerifyEmailAllowed(email, 3)
	assert.False(t, allowed)
}

func TestStoreEmailBounce(t *T) {
	require.False(t, mongoDisabled)
	emailSH.WithColl(func(c *mgo.Collection) {
		email := "test2@test.com"
		err := StoreEmailBounce(email)
		require.Nil(t, err)
		doc := &EmailDoc{}
		err = c.FindId(email).One(doc)
		require.Nil(t, err)

		assert.Equal(t, 1, len(doc.Bounces))
		//make sure bounce time is within 1 second
		diff := time.Now().Sub(doc.Bounces[0])
		assert.True(t, diff < time.Second)
	})
}

func TestStoreEmailSpam(t *T) {
	require.False(t, mongoDisabled)
	emailSH.WithColl(func(c *mgo.Collection) {
		email := "test3@test.com"
		err := StoreEmailSpam(email)
		require.Nil(t, err)
		doc := &EmailDoc{}
		err = c.FindId(email).One(doc)
		require.Nil(t, err)

		assert.Equal(t, 1, len(doc.SpamReports))
		//make sure bounce time is within 1 second
		diff := time.Now().Sub(doc.SpamReports[0])
		assert.True(t, diff < time.Second)
	})
}
