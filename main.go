package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu-community/sensu-plugin-sdk/templates"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// Config represents the handler plugin config.
type Config struct {
	sensu.PluginConfig
	APIURL       string
	StateMessage string
	EntityID     string
}

// SQEvent is the JSON type for creating a Squadcast incident
type SQEvent struct {
	MessageType    string         `json:"message_type"`
	StateMessage   string         `json:"state_message,omitempty"`
	EntityID       string         `json:"entity_id,omitempty"`
	HostName       string         `json:"host_name,omitempty"`
	MonitoringTool string         `json:"monitoring_tool,omitempty"`
	Check          *corev2.Check  `json:"check,omitempty"`
	Entity         *corev2.Entity `json:"entity,omitempty"`
}

const (
	apiurl       = "api-url"
	statemessage = "state-message"
	entityid     = "entity-id"
)

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-squadcast-handler",
			Short:    "sends  sensu  events to squadcast",
			Keyspace: "sensu.io/plugins/sensu-squadcast-handler/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      apiurl,
			Env:       "SENSU_SQUADCAST_APIURL",
			Argument:  apiurl,
			Shorthand: "a",
			Default:   "",
			Secret:    true,
			Usage:     "The URL for the Squadcast API",
			Value:     &plugin.APIURL,
		},
		{
			Path:      statemessage,
			Env:       "SENSU_SQUADCAST_STATE_MESSAGE",
			Argument:  statemessage,
			Shorthand: "s",
			Default:   "{{.Entity.Name}}:{{.Check.Name}}:{{.Check.Output}}",
			Usage:     "The template to use for the state message",
			Value:     &plugin.StateMessage,
		},
		{
			Path:      entityid,
			Env:       "SENSU_SQUADCAST_ENTITY_ID",
			Argument:  entityid,
			Shorthand: "e",
			Default:   "{{.Entity.Name}}/{{.Check.Name}}",
			Usage:     "The template to use for the Entity ID",
			Value:     &plugin.EntityID,
		},
	}
)

// CheckArgs validates the configuration passed to the handler
func CheckArgs(_ *corev2.Event) error {
	if len(plugin.APIURL) == 0 {
		return errors.New("missing Squadcast API URL")
	}
	if !govalidator.IsURL(plugin.APIURL) {
		return errors.New("invlaid Squadcast API URL specification")
	}
	return nil
}

// SendEventToSquadcast sends the event data to configured Squadcast Webhook endpoint
func SendEventToSquadcast(event *corev2.Event) error {
	var msgType string

	switch eventStatus := event.Check.Status; eventStatus {
	case 0:
		msgType = "RECOVERY"
	case 1:
		msgType = "WARNING"
	default:
		msgType = "CRITICAL"
	}
	msgEntityID, err := templates.EvalTemplate("entityID", plugin.EntityID, event)
	if err != nil {
		return fmt.Errorf("failed to evaluate template %s: %v", plugin.EntityID, err)
	}
	msgStateMessage, err := templates.EvalTemplate("stateMessage", plugin.StateMessage, event)
	if err != nil {
		return fmt.Errorf("failed to evaluate template %s: %v", plugin.StateMessage, err)
	}
	sqEvent := SQEvent{
		MessageType:    msgType,
		StateMessage:   msgStateMessage,
		EntityID:       msgEntityID,
		HostName:       event.Entity.Name,
		MonitoringTool: "sensu",
		Check:          event.Check,
		Entity:         event.Entity,
	}
	msgBytes, err := json.Marshal(sqEvent)
	if err != nil {
		return fmt.Errorf("Failed to marshal Squadcast event: %s", err)
	}

	resp, err := http.Post(plugin.APIURL, "application/json", bytes.NewBuffer(msgBytes))
	if err != nil {
		return fmt.Errorf("Post to %s failed: %s", plugin.APIURL, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("POST to %s failed with %v", plugin.APIURL, resp.Status)
	}
	return nil
}

func main() {
	handler := sensu.NewGoHandler(&plugin.PluginConfig, options, CheckArgs, SendEventToSquadcast)
	handler.Execute()
}
