// Package db manages the queuing/persistance for the postmaster
package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/levenlabs/go-llog"
	"github.com/levenlabs/golib/genapi"
	"github.com/levenlabs/postmaster/ga"
	"github.com/levenlabs/postmaster/sender"
	"github.com/mediocregopher/okq-go.v2"
)

var (
	normalQueue     = "email-normal"
	statsQueue      = "stats-normal"
	uniqueArgStatID = "pmStatsID"
	uniqueArgEnvID  = "pmEnvID"
)

var jobCh chan job

var useOkq bool

type job struct {
	Queue    string
	Contents string
	RespCh   chan error
}

func init() {
	ga.GA.AppendInit(func(g *genapi.GenAPI) {
		if ga.GA.OkqInfo.Client == nil {
			return
		}
		jobCh = make(chan job)

		okqClient := ga.GA.OkqInfo.Client
		// Receive jobs from StoreSendJob() and StoreStatsJob() and Push into okq
		go func() {
			for job := range jobCh {
				job.RespCh <- okqClient.Push(job.Queue, job.Contents, okq.Normal)
			}
		}()

		// Receive jobs from okq and send to sender
		consumeSpin(handleSendEvent, normalQueue)

		// Receive jobs from okq and store in stats
		consumeSpin(handleStatsEvent, statsQueue)

		useOkq = true
	})
}

// DisableOkq turns off using okq for job storing
// this should ONLY be called during testing
func DisableOkq() {
	useOkq = false
}

func consumeSpin(fn okq.ConsumerFunc, q string) {
	llog.Info("creating okq consumer", llog.KV{"queue": q})
	consumer := ga.GA.OkqInfo.Client
	go func(c *okq.Client) {
		for {
			err := <-c.Consumer(context.Background(), fn, q)
			llog.Error("consumer error", llog.KV{"queue": q}, llog.ErrKV(err))
			time.Sleep(10 * time.Second)
		}
	}(consumer)
}

// StoreSendJob creates a new Mail job with jobContents and sends it to okq
func StoreSendJob(jobContents string) error {
	if !useOkq {
		if !sendEmail(jobContents) {
			return errors.New("Failed to send email (bypassing okq)")
		}
		return nil
	}
	respCh := make(chan error)
	jobCh <- job{normalQueue, jobContents, respCh}
	return <-respCh
}

// StoreStatsJob creates a new statsJob with jobContents and sends it to okq
func StoreStatsJob(jobContents string) error {
	if !useOkq {
		if !storeStats(jobContents) {
			return errors.New("Failed to store stats (bypassing okq)")
		}
		return nil
	}
	respCh := make(chan error)
	jobCh <- job{statsQueue, jobContents, respCh}
	return <-respCh
}

func handleSendEvent(_ context.Context, e okq.Event) bool {
	return sendEmail(e.Contents)
}

func handleStatsEvent(_ context.Context, e okq.Event) bool {
	return storeStats(e.Contents)
}

func sendEmail(jobContents string) bool {
	job := new(sender.Mail)
	err := json.Unmarshal([]byte(jobContents), job)
	if err != nil {
		llog.Error("error json decoding into sender.Mail", llog.KV{
			"jobContents": jobContents,
		}, llog.ErrKV(err))
		// since we cannot process this job, no reason to have it keep around
		return true
	}

	env := ga.Environment
	id := GenerateEmailID(job.To, job.Flags, job.UniqueID, env)
	if id != "" {
		if job.UniqueArgs == nil {
			job.UniqueArgs = make(map[string]string)
		}
		job.UniqueArgs[uniqueArgStatID] = id
		job.UniqueArgs[uniqueArgEnvID] = env
	}

	llog.Info("processing send job", llog.KV{"id": id, "recipient": job.To})
	err = sender.Send(job)
	if err != nil {
		if id != "" {
			// if we ran into an error sending the email, delete the emailID
			rerr := removeEmailID(id)
			if rerr != nil {
				llog.Error("error deleting failed emailID", llog.KV{"id": id}, llog.ErrKV(err))
			}
		}

		llog.Error("error calling sender.Send", llog.KV{"jobContents": jobContents, "id": id}, llog.ErrKV(err))
		return false
	}
	return true
}

func logMarkError(err error, kv llog.KV) {
	if err != nil {
		llog.Error("error marking email", kv, llog.ErrKV(err))
	}
}

func storeStats(jobContents string) bool {
	job := new(StatsJob)
	err := json.Unmarshal([]byte(jobContents), job)
	if err != nil {
		llog.Error("error json decoding into StatsJob", llog.KV{
			"jobContents": jobContents,
			"err":         err,
		})
		// since we cannot process this job, no reason to have it keep around
		return true
	}

	kv := llog.KV{
		"id":     job.StatsID,
		"type":   job.Type,
		"reason": job.Reason,
		"email":  job.Email,
	}
	llog.Info("processing stats job", kv)
	switch job.Type {
	case "delivered":
		err = MarkAsDelivered(job.StatsID)
		logMarkError(err, kv)
	case "open":
		err = MarkAsOpened(job.StatsID)
		logMarkError(err, kv)
	case "bounce":
		err = MarkAsBounced(job.StatsID, job.Reason)
		logMarkError(err, kv)

		err = StoreEmailBounce(job.Email)
		if err != nil {
			llog.Error("error storing email as bounced", kv, llog.ErrKV(err))
		}
	case "spamreport":
		err = MarkAsSpamReported(job.StatsID)
		logMarkError(err, kv)

		err = StoreEmailSpam(job.Email)
		if err != nil {
			llog.Error("error storing email as spamed", kv, llog.ErrKV(err))
		}
	case "dropped":
		//depending on the reason we should mark the email as invalid
		err = MarkAsDropped(job.StatsID, job.Reason)
		logMarkError(err, kv)
	default:
		llog.Warn("received unknown job type", llog.KV{"type": job.Type})
	}
	return true
}
