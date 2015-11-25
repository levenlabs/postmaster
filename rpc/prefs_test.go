package rpc

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/validator.v2"
	. "testing"
)

func TestValidation(t *T) {
	s := &UpdatePrefsArgs{}
	assert.NotNil(t, validator.Validate(s))

	s = &UpdatePrefsArgs{
		Email: "fake",
		Flags: 1,
	}
	assert.NotNil(t, validator.Validate(s))

	s = &UpdatePrefsArgs{
		Email: "fake@gmail.com",
	}
	assert.NotNil(t, validator.Validate(s))
}
