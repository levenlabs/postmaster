package sender

import (
	"github.com/levenlabs/golib/testutil"
	"github.com/stretchr/testify/assert"
	"gopkg.in/validator.v2"
	. "testing"
)

func TestValidation(t *T) {
	s := &Mail{}
	assert.NotNil(t, validator.Validate(s))

	s = &Mail{
		To:      "fake",
		From:    "fake@gmail.com",
		Subject: testutil.RandStr(),
	}
	assert.NotNil(t, validator.Validate(s))

	s = &Mail{
		To:      "fake@gmail.com",
		From:    "fake",
		Subject: testutil.RandStr(),
	}
	assert.NotNil(t, validator.Validate(s))

	s = &Mail{
		To:   "fake@gmail.com",
		From: "fake@gmail.com",
	}
	assert.NotNil(t, validator.Validate(s))
}
