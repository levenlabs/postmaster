// Package sender manages actually sending the emails for the postmaster
package sender

import (
	"fmt"
	"github.com/levenlabs/go-llog"
	"github.com/levenlabs/golib/rpcutil"
	"github.com/levenlabs/postmaster/config"
	"github.com/sendgrid/sendgrid-go"
	"gopkg.in/validator.v2"
	"reflect"
)

var (
	client *sendgrid.SGClient
)

type Mail struct {
	To         string            `json:"to" validate:"email,nonzero,max=256"`
	ToName     string            `json:"toName,omitempty" validate:"max=256"`
	From       string            `json:"from" validate:"email,nonzero,max=256"`
	FromName   string            `json:"fromName,omitempty"	validate:"max=256"`
	Subject    string            `json:"subject" validate:"nonzero,max=998"`    // RFC 5322 says not longer than 998
	HTML       string            `json:"html,omitempty" validate:"max=2097152"` //2MB
	Text       string            `json:"text,omitempty" validate:"max=2097152"` //2MB
	ReplyTo    string            `json:"replyTo,omitempty" validate:"max=256"`
	UniqueArgs map[string]string `json:"uniqueArgs,omitempty" validate:"argsMap=max=256"`
	Flags      int64             `json:"flags"`
}

func init() {
	if config.SendGridAPIKey == "" {
		llog.Fatal("--sendgrid-key not set")
	}
	client = sendgrid.NewSendGridClientWithApiKey(config.SendGridAPIKey)

	rpcutil.InstallCustomValidators()
	validator.SetValidationFunc("argsMap", validateArgsMap)
}

func Send(job *Mail) error {
	msg := sendgrid.NewMail()
	msg.AddTo(job.To)
	if job.ToName != "" {
		msg.AddToName(job.ToName)
	}
	msg.SetFrom(job.From)
	if job.FromName != "" {
		msg.SetFromName(job.FromName)
	}
	msg.SetSubject(job.Subject)
	if job.HTML != "" {
		msg.SetHTML(job.HTML)
	}
	if job.Text != "" {
		msg.SetText(job.Text)
	}
	if job.ReplyTo != "" {
		msg.SetReplyTo(job.ReplyTo)
	}
	if job.UniqueArgs != nil && len(job.UniqueArgs) > 0 {
		msg.SMTPAPIHeader.SetUniqueArgs(job.UniqueArgs)
	}
	return client.Send(msg)
}

// validateArgsMap maps over the args map and validates each key and value in
// it using the passed in tag
func validateArgsMap(v interface{}, param string) error {
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}

	if k := vv.Kind(); k != reflect.Map {
		return fmt.Errorf("non-array type: %s", k)
	}

	ks := vv.MapKeys()
	for _, k := range ks {
		//first check the key
		if err := validator.Valid(k.Interface(), param); err != nil {
			return fmt.Errorf("invalid key %s", k.String(), err)
		}
		//now check the value
		kv := vv.MapIndex(k).Interface()
		if err := validator.Valid(kv, param); err != nil {
			return fmt.Errorf("invalid value at key %s: %s", k.String(), err)
		}
	}
	return nil
}
