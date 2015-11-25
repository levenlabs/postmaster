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
	UniqueArgID = "stats_id"
)

var jobCh chan job = make(chan job)

type job struct {
	Queue    string
	Contents string
	RespCh   chan error
}

func init() {
	if config.OKQAddr == "" {
		return
	}

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
	okqSendConsumer := okq.New(config.OKQAddr)
	// Receive jobs from okq and send to sender
	go func() {
		for {
			err := okqSendConsumer.Consumer(handleSendEvent, nil, normalQueue)
			llog.Error("send consumer error", llog.KV{"err": err})
			time.Sleep(10 * time.Second)
		}
	}()

	llog.Info("creating okq stats consumer", llog.KV{"okqAddr": config.OKQAddr})
	okqStatsConsumer := okq.New(config.OKQAddr)
	// Receive jobs from okq and store in stats
	go func() {
		for {
			err := okqStatsConsumer.Consumer(handleStatsEvent, nil, statsQueue)
			llog.Error("stats consumer error", llog.KV{"err": err})
			time.Sleep(10 * time.Second)
		}
	}()
}

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
		llog.Error("error json decoding into sender.Mail", llog.KV{"jobContents": jobContents, "err": err})
		return false
	}

	id := GenerateEmailID(job.To, job.Flags)
	if id != "" {
		if job.UniqueArgs == nil {
			job.UniqueArgs = make(map[string]string)
		}
		job.UniqueArgs[UniqueArgID] = id
	}

	llog.Info("processing send job", llog.KV{"id": id, "recipient": job.To})
	err = sender.Send(job)
	if err != nil {
		llog.Error("error calling sender.Send", llog.KV{"jobContents": jobContents, "err": err})
		return false
	}
	return true
}

func storeStats(jobContents string) bool {
	job := new(StatsJob)
	err := json.Unmarshal([]byte(jobContents), job)
	if err != nil {
		llog.Error("error json decoding into StatsJob", llog.KV{"jobContents": jobContents, "err": err})
		return false
	}

	llog.Info("processing stats job", llog.KV{"job": job})
	switch job.Type {
	case "delivered":
		MarkAsDelivered(job.StatsID)
	case "open":
		MarkAsOpened(job.StatsID)
	case "bounce":
		MarkAsBounced(job.StatsID, job.Reason)
		StoreEmailBounce(job.Email)
	case "spamreport":
		MarkAsSpamReported(job.StatsID)
		StoreEmailSpam(job.Email)
	case "dropped":
		//depending on the reason we should mark the email as invalid
		MarkAsDropped(job.StatsID, job.Reason)
	default:
		llog.Warn("received unknown job type", llog.KV{"type": job.Type})
	}
	return true
}
