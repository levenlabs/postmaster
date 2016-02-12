// Package db manages the queuing/persistance for the postmaster
package db

import (
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"github.com/levenlabs/go-llog"
	"github.com/levenlabs/postmaster/config"
	"github.com/levenlabs/postmaster/sender"
	"github.com/mediocregopher/okq-go/okq"
)

var (
	normalQueue = "email-normal"
	statsQueue  = "stats-normal"
	uniqueArgID = "pmStatsID"
)

var jobCh chan job

type job struct {
	Queue    string
	Contents string
	RespCh   chan error
}

func init() {
	if config.OKQAddr == "" {
		return
	}
	jobCh = make(chan job)

	rand.Seed(time.Now().UTC().UnixNano())

	llog.Info("creating okq client", llog.KV{"okqAddr": config.OKQAddr})
	okqClient := okq.New(config.OKQAddr)
	// Receive jobs from StoreSendJob() and StoreStatsJob() and Push into okq
	go func() {
		for job := range jobCh {
			job.RespCh <- okqClient.Push(job.Queue, job.Contents, okq.Normal)
		}
	}()

	llog.Info("creating okq send consumer", llog.KV{"okqAddr": config.OKQAddr})
	// Receive jobs from okq and send to sender
	consumeSpin(handleSendEvent, normalQueue)

	llog.Info("creating okq stats consumer", llog.KV{"okqAddr": config.OKQAddr})
	// Receive jobs from okq and store in stats
	consumeSpin(handleStatsEvent, statsQueue)
}

func consumeSpin(fn func(e *okq.Event) bool, q string) {
	consumer := okq.New(config.OKQAddr)
	go func(c *okq.Client) {
		for {
			err := c.Consumer(fn, nil, q)
			llog.Error("consumer error", llog.KV{"err": err, "queue": q})
			time.Sleep(10 * time.Second)
		}
	}(consumer)
}

// StoreSendJob creates a new Mail job with jobContents and sends it to okq
func StoreSendJob(jobContents string) error {
	if jobCh == nil {
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
	if jobCh == nil {
		if !storeStats(jobContents) {
			return errors.New("Failed to store stats (bypassing okq)")
		}
		return nil
	}
	respCh := make(chan error)
	jobCh <- job{statsQueue, jobContents, respCh}
	return <-respCh
}

func handleSendEvent(e *okq.Event) bool {
	return sendEmail(e.Contents)
}

func handleStatsEvent(e *okq.Event) bool {
	return storeStats(e.Contents)
}

func sendEmail(jobContents string) bool {
	job := new(sender.Mail)
	err := json.Unmarshal([]byte(jobContents), job)
	if err != nil {
		llog.Error("error json decoding into sender.Mail", llog.KV{
			"jobContents": jobContents,
			"err":         err,
		})
		// since we cannot process this job, no reason to have it keep around
		return true
	}

	id := GenerateEmailID(job.To, job.Flags, job.UniqueID)
	if id != "" {
		if job.UniqueArgs == nil {
			job.UniqueArgs = make(map[string]string)
		}
		job.UniqueArgs[uniqueArgID] = id
	}

	llog.Info("processing send job", llog.KV{"id": id, "recipient": job.To})
	err = sender.Send(job)
	if err != nil {
		if id != "" {
			// if we ran into an error sending the email, delete the emailID
			rerr := removeEmailID(id)
			if rerr != nil {
				llog.Error("error deleting failed emailID",
					llog.KV{"id": id, "err": err})
			}
		}

		llog.Error("error calling sender.Send", llog.KV{"jobContents": jobContents, "id": id, "err": err})
		return false
	}
	return true
}

func logMarkError(err error, kv llog.KV) {
	if err != nil {
		kv["error"] = err
		llog.Error("error marking email", kv)
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
			llog.Error("error storing email as bounced", kv)
		}
	case "spamreport":
		err = MarkAsSpamReported(job.StatsID)
		logMarkError(err, kv)

		err = StoreEmailSpam(job.Email)
		if err != nil {
			llog.Error("error storing email as spamed", kv)
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
