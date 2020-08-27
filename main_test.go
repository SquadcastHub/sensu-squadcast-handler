package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckArgs(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	assert.Error(CheckArgs(event))
	plugin.APIURL = "InvalidURL"
	assert.Error(CheckArgs(event))
	plugin.APIURL = "http://sensu.example.com:3000"
	assert.NoError(CheckArgs(event))
}

func TestSendEventToSquadcast(t *testing.T) {
	testcases := []struct {
		status  uint32
		msgtype string
	}{
		{0, "RECOVERY"},
		{1, "WARNING"},
		{2, "CRITICAL"},
		{127, "CRITICAL"},
	}

	for _, tc := range testcases {
		assert := assert.New(t)
		event := corev2.FixtureEvent("entity1", "check1")
		event.Check.Status = tc.status

		var test = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(err)
			msg := &SQEvent{}
			err = json.Unmarshal(body, msg)
			require.NoError(t, err)
			expectedEntityID := "entity1/check1"
			assert.Equal(expectedEntityID, msg.EntityID)
			expectedMessageType := tc.msgtype
			assert.Equal(expectedMessageType, msg.MessageType)
			w.WriteHeader(http.StatusOK)
		}))

		_, err := url.ParseRequestURI(test.URL)
		require.NoError(t, err)
		plugin.APIURL = test.URL
		assert.NoError(SendEventToSquadcast(event))
	}
}
