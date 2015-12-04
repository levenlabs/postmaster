package rpc

import (
	"github.com/levenlabs/postmaster/sender"
	"github.com/stretchr/testify/assert"
	. "testing"
)

func TestValidation(t *T) {
	// this should fail because it ends in @test
	a := &sender.Mail{
		To:   "test@test",
		Text: "hey",
	}
	assert.NotNil(t, validateEnqueueArgs(a))

	// this should fail because it has no HTML or Text
	a = &sender.Mail{
		To: "test@gmail.com",
	}
	assert.NotNil(t, validateEnqueueArgs(a))

	// this should not fail because it has no HTML or Text
	a = &sender.Mail{
		To:   "test@gmail.com",
		Text: "hey",
	}
	assert.Nil(t, validateEnqueueArgs(a))
}
