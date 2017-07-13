// Package sender manages actually sending the emails for the postmaster
package sender

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/levenlabs/go-llog"
	"github.com/levenlabs/golib/genapi"
	"github.com/levenlabs/golib/rpcutil"
	"github.com/levenlabs/postmaster/ga"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"gopkg.in/validator.v2"
)

var (
	sgKey  string
	sgPool string
)

// Mail encompasses an email that is intended to be sent
type Mail struct {
	// To is the email address of the recipient
	To string `json:"to" validate:"email,nonzero,max=256"`

	// ToName is optional and represents the recipeient's name
	ToName string `json:"toName,omitempty" validate:"max=256"`

	// From is the email address of the sender
	From string `json:"from" validate:"email,nonzero,max=256"`

	// FromName is optional and represents the name of the sender
	FromName string `json:"fromName,omitempty" validate:"max=256"`

	// Subject is the subject of the email
	Subject string `json:"subject" validate:"nonzero,max=998"` // RFC 5322 says not longer than 998

	// HTML is the HTML body of the email and is required unless Text is sent
	HTML string `json:"html,omitempty" validate:"max=2097152"` //2MB

	// Text is the plain-text body and is required unless HTML is sent
	Text string `json:"text,omitempty" validate:"max=2097152"` //2MB

	// ReplyTo is the Reply-To email address for the email
	ReplyTo string `json:"replyTo,omitempty" validate:"email,max=256"`

	// UniqueArgs are the SMTP unique arguments passed onto sendgrid
	// Note: pmStatsID is a reserved key and is used for stats recording
	UniqueArgs map[string]string `json:"uniqueArgs,omitempty" validate:"argsMap=max=256"`

	// Flags represent the category flags for this email and are used to
	// determine if the recipient has blocked this category of email
	Flags int64 `json:"flags"`

	// UniqueID is an optional uniqueID for this email that will be stored with
	// the email stats and can be used to later query when the last email with
	// this ID was sent
	UniqueID string `json:"uniqueID,omitempty" validate:"max=256"`
}

func init() {
	ga.GA.AppendInit(func(g *genapi.GenAPI) {
		key, _ := g.ParamStr("--sendgrid-key")
		if key == "" {
			llog.Fatal("--sendgrid-key not set")
		}
		sgKey = key
		sgPool, _ = g.ParamStr("--sendgrid-ip-pool")

		rpcutil.InstallCustomValidators()
		validator.SetValidationFunc("argsMap", validateArgsMap)
	})
}

// Send takes a Mail struct and sends it to sendgrid
func Send(job *Mail) error {
	msg := mail.NewV3Mail()
	msg.SetFrom(mail.NewEmail(job.FromName, job.From))
	if job.ReplyTo != "" {
		msg.SetReplyTo(mail.NewEmail("", job.ReplyTo))
	}

	p := mail.NewPersonalization()
	// make sure the To doesn't have any <> in it since this breaks sendgrid
	tn := strings.Replace(job.ToName, "<", "", -1)
	tn = strings.Replace(tn, ">", "", -1)
	p.AddTos(mail.NewEmail(tn, job.To))
	msg.AddPersonalizations(p)

	msg.Subject = job.Subject
	contents := []*mail.Content{}
	if job.HTML != "" {
		contents = append(contents, mail.NewContent("text/html", job.HTML))
	}
	if job.Text != "" {
		contents = append(contents, mail.NewContent("text/plain", job.Text))
	}
	msg.AddContent(contents...)

	if job.UniqueArgs != nil && len(job.UniqueArgs) > 0 {
		for k, v := range job.UniqueArgs {
			msg.SetCustomArg(k, v)
		}
	}
	if sgPool != "" {
		msg.SetIPPoolID(sgPool)
	}
	req := sendgrid.GetRequest(sgKey, "/v3/mail/send", "https://api.sendgrid.com")
	req.Method = "POST"
	req.Body = mail.GetRequestBody(msg)
	resp, err := sendgrid.API(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusAccepted {
		return errors.New(resp.Body)
	}
	return nil
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
			return fmt.Errorf("invalid key %s: %s", k.String(), err)
		}
		//now check the value
		kv := vv.MapIndex(k).Interface()
		if err := validator.Valid(kv, param); err != nil {
			return fmt.Errorf("invalid value at key %s: %s", k.String(), err)
		}
	}
	return nil
}
