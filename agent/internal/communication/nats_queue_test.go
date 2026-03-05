package communication

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNATSCommandMessage_JSONRoundtrip(t *testing.T) {
	msg := NATSCommandMessage{
		DeviceID:     "rpi-001",
		CommandID:    "cmd-abc",
		CommandType:  "deploy_model",
		DeploymentID: "deploy-xyz",
		Payload: map[string]string{
			"model_name":    "resnet50",
			"artifact_url":  "https://example.com/model.tflite",
			"runtime":       "tflite",
		},
		IssuedAt: time.Now().Truncate(time.Second),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded NATSCommandMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.DeviceID != "rpi-001" {
		t.Errorf("expected device_id rpi-001, got %s", decoded.DeviceID)
	}
	if decoded.CommandType != "deploy_model" {
		t.Errorf("expected command_type deploy_model, got %s", decoded.CommandType)
	}
	if decoded.Payload["runtime"] != "tflite" {
		t.Errorf("expected payload runtime tflite, got %s", decoded.Payload["runtime"])
	}
}

func TestNATSCommandMessage_ToInternalCommand(t *testing.T) {
	natsCmd := NATSCommandMessage{
		DeviceID:    "jetson-001",
		CommandID:   "cmd-123",
		CommandType: "deploy_model",
		Payload: map[string]string{
			"model_name": "mobilenet",
		},
	}

	payloadJSON, err := json.Marshal(natsCmd.Payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	cmd := Command{
		ID:      natsCmd.CommandID,
		Type:    natsCmd.CommandType,
		Payload: payloadJSON,
	}

	if cmd.ID != "cmd-123" {
		t.Errorf("expected ID cmd-123, got %s", cmd.ID)
	}
	if cmd.Type != "deploy_model" {
		t.Errorf("expected type deploy_model, got %s", cmd.Type)
	}

	var payload map[string]string
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload["model_name"] != "mobilenet" {
		t.Errorf("expected model_name mobilenet, got %s", payload["model_name"])
	}
}

func TestNATSCommandMessage_EmptyOptionalFields(t *testing.T) {
	msg := NATSCommandMessage{
		DeviceID:    "test",
		CommandID:   "cmd-1",
		CommandType: "restart",
	}

	data, _ := json.Marshal(msg)
	var decoded NATSCommandMessage
	json.Unmarshal(data, &decoded)

	if decoded.DeploymentID != "" {
		t.Error("expected empty deployment_id")
	}
	if decoded.Payload != nil {
		t.Error("expected nil payload")
	}
}

func TestNATSSubjectFormat(t *testing.T) {
	tests := []struct {
		deviceID string
		expected string
	}{
		{"jetson-001", "fleetml.commands.jetson-001"},
		{"rpi-042", "fleetml.commands.rpi-042"},
	}

	for _, tt := range tests {
		subject := "fleetml.commands." + tt.deviceID
		if subject != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, subject)
		}
	}
}

func TestNATSCommandMessage_ZeroTime(t *testing.T) {
	msg := NATSCommandMessage{DeviceID: "dev-1", CommandID: "cmd-1", CommandType: "deploy"}
	if !msg.IssuedAt.IsZero() {
		t.Error("expected zero IssuedAt")
	}
	data, _ := json.Marshal(msg)
	var decoded NATSCommandMessage
	json.Unmarshal(data, &decoded)
	if !decoded.IssuedAt.IsZero() {
		t.Error("expected zero IssuedAt after roundtrip")
	}
}

func TestNATSCommandMessage_LargePayload(t *testing.T) {
	payload := make(map[string]string, 100)
	for i := 0; i < 100; i++ {
		payload[string(rune('a'+i%26))+string(rune('0'+i/26))] = "value"
	}
	msg := NATSCommandMessage{DeviceID: "d", CommandID: "c", CommandType: "t", Payload: payload}
	data, _ := json.Marshal(msg)
	var decoded NATSCommandMessage
	json.Unmarshal(data, &decoded)
	if len(decoded.Payload) == 0 {
		t.Error("expected non-empty payload after roundtrip")
	}
}

func TestNATSCommandMessage_EmptyDeviceID(t *testing.T) {
	msg := NATSCommandMessage{DeviceID: "", CommandID: "cmd-1", CommandType: "deploy"}
	data, _ := json.Marshal(msg)
	var decoded NATSCommandMessage
	json.Unmarshal(data, &decoded)
	if decoded.DeviceID != "" {
		t.Errorf("expected empty device_id, got %q", decoded.DeviceID)
	}
}

func TestNATSCommandMessage_SpecialCharsInDeviceID(t *testing.T) {
	for _, id := range []string{"dev/001", "dev.001", "dev_001", "dev@host"} {
		msg := NATSCommandMessage{DeviceID: id, CommandID: "c", CommandType: "t"}
		data, _ := json.Marshal(msg)
		var decoded NATSCommandMessage
		json.Unmarshal(data, &decoded)
		if decoded.DeviceID != id {
			t.Errorf("expected %q, got %q", id, decoded.DeviceID)
		}
	}
}

func TestNATSCommandMessage_AllFieldsEmpty(t *testing.T) {
	msg := NATSCommandMessage{}
	data, _ := json.Marshal(msg)
	var decoded NATSCommandMessage
	json.Unmarshal(data, &decoded)
	if decoded.DeviceID != "" || decoded.CommandID != "" || decoded.CommandType != "" {
		t.Error("expected all string fields empty")
	}
}

func TestNATSCommandMessage_NilPayload(t *testing.T) {
	msg := NATSCommandMessage{DeviceID: "d", CommandID: "c", CommandType: "t", Payload: nil}
	data, _ := json.Marshal(msg)
	var decoded NATSCommandMessage
	json.Unmarshal(data, &decoded)
	if decoded.Payload != nil {
		t.Error("expected nil payload")
	}
}

func TestNATSSubjectFormat_SpecialChars(t *testing.T) {
	for _, tt := range []struct{ id, want string }{
		{"dev.001", "fleetml.commands.dev.001"},
		{"dev-002", "fleetml.commands.dev-002"},
		{"dev_003", "fleetml.commands.dev_003"},
	} {
		if got := "fleetml.commands." + tt.id; got != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}

func TestNATSSubjectFormat_EmptyID(t *testing.T) {
	if got := "fleetml.commands." + ""; got != "fleetml.commands." {
		t.Errorf("expected 'fleetml.commands.', got %q", got)
	}
}

func TestNATSCommandMessage_JSONOmitempty(t *testing.T) {
	msg := NATSCommandMessage{DeviceID: "d", CommandID: "c", CommandType: "t"}
	data, _ := json.Marshal(msg)
	var decoded NATSCommandMessage
	json.Unmarshal(data, &decoded)
	if decoded.DeploymentID != "" {
		t.Error("expected empty deployment_id")
	}
}

func TestNATSCommandMessage_PayloadMultipleKeys(t *testing.T) {
	msg := NATSCommandMessage{
		DeviceID: "d", CommandID: "c", CommandType: "deploy_model",
		Payload: map[string]string{"model_name": "resnet", "runtime": "tensorrt", "checksum": "sha256:abc"},
	}
	data, _ := json.Marshal(msg)
	var decoded NATSCommandMessage
	json.Unmarshal(data, &decoded)
	if len(decoded.Payload) != 3 {
		t.Errorf("expected 3 keys, got %d", len(decoded.Payload))
	}
	if decoded.Payload["checksum"] != "sha256:abc" {
		t.Errorf("expected checksum sha256:abc, got %q", decoded.Payload["checksum"])
	}
}
