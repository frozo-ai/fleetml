package messaging

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCommandMessage_JSONRoundtrip(t *testing.T) {
	msg := CommandMessage{
		DeviceID:     "jetson-001",
		CommandID:    "cmd-123",
		CommandType:  "deploy_model",
		DeploymentID: "deploy-456",
		Payload: map[string]string{
			"model_name":    "mobilenet",
			"model_version": "v2",
			"artifact_url":  "https://s3.example.com/model.onnx",
		},
		IssuedAt: time.Now().Truncate(time.Second),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded CommandMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.DeviceID != "jetson-001" {
		t.Errorf("expected device_id jetson-001, got %s", decoded.DeviceID)
	}
	if decoded.CommandType != "deploy_model" {
		t.Errorf("expected command_type deploy_model, got %s", decoded.CommandType)
	}
	if decoded.Payload["model_name"] != "mobilenet" {
		t.Errorf("expected payload model_name mobilenet, got %s", decoded.Payload["model_name"])
	}
}

func TestCommandMessage_SubjectFormat(t *testing.T) {
	tests := []struct {
		deviceID string
		expected string
	}{
		{"jetson-001", "fleetml.commands.jetson-001"},
		{"rpi-042", "fleetml.commands.rpi-042"},
		{"intel-nuc-abc", "fleetml.commands.intel-nuc-abc"},
	}

	for _, tt := range tests {
		subject := "fleetml.commands." + tt.deviceID
		if subject != tt.expected {
			t.Errorf("expected subject %s, got %s", tt.expected, subject)
		}
	}
}

func TestCommandMessage_EmptyPayload(t *testing.T) {
	msg := CommandMessage{
		DeviceID:    "test-device",
		CommandID:   "cmd-1",
		CommandType: "restart",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded CommandMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Payload != nil {
		t.Error("expected nil payload for restart command")
	}
	if decoded.DeploymentID != "" {
		t.Error("expected empty deployment_id for restart command")
	}
}

func TestCommandMessage_AllCommandTypes(t *testing.T) {
	types := []string{"deploy_model", "rollback", "update_config", "restart", "set_ab_test"}
	for _, ct := range types {
		msg := CommandMessage{
			DeviceID:    "test",
			CommandID:   "cmd-1",
			CommandType: ct,
		}

		data, _ := json.Marshal(msg)
		var decoded CommandMessage
		json.Unmarshal(data, &decoded)

		if decoded.CommandType != ct {
			t.Errorf("expected command_type %s, got %s", ct, decoded.CommandType)
		}
	}
}
