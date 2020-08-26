package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// Config represents the handler plugin config.
type Config struct {
	sensu.PluginConfig
	APIURL string
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

const apiurl = "api-url"

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
			Usage:     "The URL for the Squadcast API",
			Value:     &plugin.APIURL,
		},
	}
)

func main() {
	handler := sensu.NewGoHandler(&plugin.PluginConfig, options, CheckArgs, SendSquadcast)
	handler.Execute()
}

func CheckArgs(_ *corev2.Event) error {
	if len(plugin.APIURL) == 0 {
		return errors.New("missing Squadcast API URL")
	}
	if !govalidator.IsURL(plugin.APIURL) {
		return errors.New("invlaid Squadcast API URL specification")
	}
	return nil
}

func SendSquadcast(event *corev2.Event) error {
	var msgType string

	switch eventStatus := event.Check.Status; eventStatus {
	case 0:
		msgType = "RECOVERY"
	case 1:
		msgType = "WARNING"
	default:
		msgType = "CRITICAL"
	}
	msgEntityID := fmt.Sprintf("%s/%s", event.Entity.Name, event.Check.Name)
	msgStateMessage := fmt.Sprintf("%s:%s:%s", event.Entity.Name, event.Check.Name, event.Check.Output)
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
