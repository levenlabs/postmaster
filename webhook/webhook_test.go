package webhook

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	. "testing"

	"github.com/levenlabs/postmaster/config"
	"github.com/levenlabs/postmaster/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	db.RandomizeColls()
	db.DisableOkq()
}

var testEmail = "webhooktest@test"

func TestHookHandlerPassword(t *T) {
	config.WebhookPassword = "test"

	str := []byte(`[{"email":"webhooktest@test","timestamp":1,"event":"test","pmStatsID":"s"}]`)
	r, _ := http.NewRequest("POST", "/", bytes.NewBuffer(str))
	r.Header.Set("Content-Type", "application/json")
	r.SetBasicAuth("anything", "test")
	w := httptest.NewRecorder()

	hookHandler(w, r)
	assert.Equal(t, 200, w.Code)
}

func TestHookHandlerOpen(t *T) {
	config.WebhookPassword = ""

	id := db.GenerateEmailID(testEmail, 0, "", "production")
	str := []byte(fmt.Sprintf(`[{"email":"webhooktest@test","timestamp":1449264108,"pmStatsID":"%s","pmEnvID":"production","event":"open"}]`, id))
	r, _ := http.NewRequest("POST", "/", bytes.NewBuffer(str))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	hookHandler(w, r)
	assert.Equal(t, 200, w.Code)

	doc, err := db.GetStats(id)
	require.Nil(t, err)
	assert.Equal(t, int64(db.Opened), doc.StateFlags)
}

func TestHookHandlerDelivered(t *T) {
	config.WebhookPassword = ""

	id := db.GenerateEmailID(testEmail, 0, "", "production")
	str := []byte(fmt.Sprintf(`[{"email":"webhooktest@test","timestamp":1449264108,"pmStatsID":"%s","pmEnvID":"production","event":"delivered"}]`, id))
	r, _ := http.NewRequest("POST", "/", bytes.NewBuffer(str))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	hookHandler(w, r)
	assert.Equal(t, 200, w.Code)

	doc, err := db.GetStats(id)
	require.Nil(t, err)
	assert.Equal(t, int64(db.Delivered), doc.StateFlags)
}

func TestHookHandlerDropped(t *T) {
	config.WebhookPassword = ""

	id := db.GenerateEmailID("webhooktest@test.com", 0, "", "production")
	str := []byte(fmt.Sprintf(`[{"email":"webhooktest@test","timestamp":1449264108,"pmStatsID":"%s","pmEnvID":"production","event":"dropped","reason":"Test"}]`, id))
	r, _ := http.NewRequest("POST", "/", bytes.NewBuffer(str))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	hookHandler(w, r)
	assert.Equal(t, 200, w.Code)

	doc, err := db.GetStats(id)
	require.Nil(t, err)
	assert.Equal(t, int64(db.Dropped), doc.StateFlags)
	assert.Equal(t, "Test", doc.Error)
}

func TestHookHandlerBounced(t *T) {
	config.WebhookPassword = ""

	id := db.GenerateEmailID("webhooktest@test.com", 0, "", "production")
	str := []byte(fmt.Sprintf(`[{"email":"webhooktest@test","timestamp":1449264108,"pmStatsID":"%s","pmEnvID":"production","event":"bounce","reason":"Test"}]`, id))
	r, _ := http.NewRequest("POST", "/", bytes.NewBuffer(str))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	hookHandler(w, r)
	assert.Equal(t, 200, w.Code)

	doc, err := db.GetStats(id)
	require.Nil(t, err)
	assert.Equal(t, int64(db.Bounced), doc.StateFlags)
	assert.Equal(t, "Test", doc.Error)
}

func TestHookHandlerSpamReport(t *T) {
	config.WebhookPassword = ""

	id := db.GenerateEmailID(testEmail, 0, "", "production")
	str := []byte(fmt.Sprintf(`[{"email":"webhooktest@test","timestamp":1449264108,"pmStatsID":"%s","pmEnvID":"production","event":"spamreport"}]`, id))
	r, _ := http.NewRequest("POST", "/", bytes.NewBuffer(str))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	hookHandler(w, r)
	assert.Equal(t, 200, w.Code)

	doc, err := db.GetStats(id)
	require.Nil(t, err)
	assert.Equal(t, int64(db.SpamReported), doc.StateFlags)
}

func TestHookHandlerDeliveredMultiple(t *T) {
	config.WebhookPassword = ""

	id := db.GenerateEmailID(testEmail, 0, "", "production")
	id2 := db.GenerateEmailID(testEmail, 0, "", "production")
	str := []byte(fmt.Sprintf(`[
	{"email":"webhooktest@test","timestamp":1449264108,"pmStatsID":"%s","pmEnvID":"production","event":"delivered"},
	{"email":"webhooktest@test","timestamp":1449264109,"pmStatsID":"%s","pmEnvID":"production","event":"delivered"}
	]`, id, id2))
	r, _ := http.NewRequest("POST", "/", bytes.NewBuffer(str))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	hookHandler(w, r)
	assert.Equal(t, 200, w.Code)

	doc, err := db.GetStats(id)
	require.Nil(t, err)
	assert.Equal(t, int64(db.Delivered), doc.StateFlags)

	doc, err = db.GetStats(id)
	require.Nil(t, err)
	assert.Equal(t, int64(db.Delivered), doc.StateFlags)
}

func TestHookHandlerDev(t *T) {
	config.WebhookPassword = ""

	id := db.GenerateEmailID(testEmail, 0, "", "dev")
	str := []byte(fmt.Sprintf(`[{"email":"webhooktest@test","timestamp":1449264108,"pmStatsID":"%s","pmEnvID":"dev","event":"delivered"}]`, id))
	r, _ := http.NewRequest("POST", "/", bytes.NewBuffer(str))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	hookHandler(w, r)
	assert.Equal(t, 200, w.Code)

	doc, err := db.GetStats(id)
	require.Nil(t, err)
	assert.Equal(t, int64(0), doc.StateFlags)
}
