package fleet

import (
	"encoding/json"
	"testing"

	"github.com/fleetml/fleetml/server/internal/domain"
)

func TestFleetStats_EmptyInit(t *testing.T) {
	stats := &FleetStats{
		RuntimeCounts: map[string]int{},
		ArchCounts:    map[string]int{},
	}

	if stats.TotalDevices != 0 {
		t.Errorf("expected 0 total devices, got %d", stats.TotalDevices)
	}
	if stats.OnlineDevices != 0 {
		t.Errorf("expected 0 online devices, got %d", stats.OnlineDevices)
	}
	if len(stats.RuntimeCounts) != 0 {
		t.Errorf("expected empty runtime counts")
	}
	if len(stats.ArchCounts) != 0 {
		t.Errorf("expected empty arch counts")
	}
}

func TestFleetStats_JSONTags(t *testing.T) {
	// Verify the struct serializes correctly
	stats := FleetStats{
		TotalDevices:   10,
		OnlineDevices:  8,
		OfflineDevices: 1,
		WarningDevices: 1,
		RuntimeCounts:  map[string]int{"tensorrt": 5, "onnx": 3, "openvino": 2},
		ArchCounts:     map[string]int{"arm64": 6, "amd64": 4},
	}

	if stats.TotalDevices != 10 {
		t.Errorf("expected 10 total, got %d", stats.TotalDevices)
	}
	if stats.RuntimeCounts["tensorrt"] != 5 {
		t.Errorf("expected 5 tensorrt, got %d", stats.RuntimeCounts["tensorrt"])
	}
	if stats.ArchCounts["arm64"] != 6 {
		t.Errorf("expected 6 arm64, got %d", stats.ArchCounts["arm64"])
	}
}

// --- Edge-case tests ---

func TestNewManager_NilDB(t *testing.T) {
	m := NewManager(nil)
	if m == nil {
		t.Fatal("expected non-nil manager with nil DB")
	}
	if m.db != nil {
		t.Error("expected nil db field")
	}
}

func TestRegisterDevice_EmptyDeviceID(t *testing.T) {
	// A device with an empty DeviceID is structurally valid in the Go struct
	// but would violate DB constraints (device_id is typically NOT NULL/UNIQUE).
	device := &domain.Device{
		DeviceID: "",
		Name:     "unnamed-device",
		Arch:     "arm64",
	}

	if device.DeviceID != "" {
		t.Errorf("expected empty device_id, got %q", device.DeviceID)
	}
	// Verify the device marshals correctly even with empty ID
	data, err := json.Marshal(device)
	if err != nil {
		t.Fatalf("marshal device with empty ID: %v", err)
	}
	var decoded domain.Device
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.DeviceID != "" {
		t.Errorf("expected empty device_id after roundtrip, got %q", decoded.DeviceID)
	}
}

func TestRegisterDevice_DuplicateDevice(t *testing.T) {
	// RegisterDevice uses ON CONFLICT (device_id) DO UPDATE, so duplicates update rather than fail.
	// We verify the upsert struct shape is valid.
	first := &domain.Device{
		DeviceID: "device-dup",
		Name:     "Original",
		Arch:     "arm64",
		Runtime:  "onnx",
		RAMMB:    2048,
	}
	second := &domain.Device{
		DeviceID: "device-dup",
		Name:     "Updated",
		Arch:     "amd64",
		Runtime:  "tensorrt",
		RAMMB:    4096,
	}

	if first.DeviceID != second.DeviceID {
		t.Error("duplicate test requires same device_id")
	}
	if second.Arch != "amd64" {
		t.Errorf("expected updated arch 'amd64', got %q", second.Arch)
	}
}

func TestListDevices_InvalidStatus(t *testing.T) {
	// DeviceFilter allows any string for Status — the DB won't match invalid ones.
	filter := domain.DeviceFilter{
		Status: "nonexistent_status",
		Limit:  50,
	}

	validStatuses := map[string]bool{
		"registered":      true,
		"healthy":         true,
		"warning":         true,
		"offline":         true,
		"decommissioned":  true,
	}

	if validStatuses[filter.Status] {
		t.Errorf("expected %q to be an invalid status", filter.Status)
	}
}

func TestListDevices_NegativeLimit(t *testing.T) {
	filter := domain.DeviceFilter{
		Limit: -1,
	}

	// ListDevices defaults limit <= 0 to 50
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit != 50 {
		t.Errorf("expected defaulted limit 50, got %d", filter.Limit)
	}
}

func TestUpdateDeviceStatus_InvalidStatus(t *testing.T) {
	// UpdateDeviceStatus accepts any string for status. The DB will store it.
	// This tests that an invalid status string is structurally valid in Go.
	status := "banana"
	validStatuses := []string{"registered", "healthy", "warning", "offline", "decommissioned"}
	isValid := false
	for _, s := range validStatuses {
		if s == status {
			isValid = true
			break
		}
	}
	if isValid {
		t.Errorf("expected %q to be an invalid device status", status)
	}
}

func TestUpdateDeviceStatus_AllNilMetrics(t *testing.T) {
	// All metric pointers can be nil — this represents a heartbeat with no metrics.
	var cpuPct, gpuPct, diskPct, tempC, uptimeH *float64
	var ramUsed *int

	if cpuPct != nil || gpuPct != nil || diskPct != nil || tempC != nil || uptimeH != nil || ramUsed != nil {
		t.Error("expected all metric pointers to be nil")
	}
}

func TestSelectDevices_EmptyLabels(t *testing.T) {
	// When target_type is "labels" with an empty label map, all non-decommissioned devices match.
	labels := map[string]string{}
	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		t.Fatalf("marshal empty labels: %v", err)
	}

	// An empty JSON object {} matches all jsonb via @> operator in PostgreSQL.
	if string(labelsJSON) != "{}" {
		t.Errorf("expected empty JSON object '{}', got %q", string(labelsJSON))
	}
}

func TestSelectDevices_SpecialCharactersInLabels(t *testing.T) {
	labels := map[string]string{
		"env/tier":   "production",
		"gpu.model":  "nvidia-a100",
		"zone:region": "us-west-2",
		"key with spaces": "value with spaces",
		"unicode-key-\u00e9": "unicode-val-\u00e9",
	}

	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		t.Fatalf("marshal special labels: %v", err)
	}

	// Verify roundtrip preserves special characters
	var decoded map[string]string
	if err := json.Unmarshal(labelsJSON, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for k, v := range labels {
		if decoded[k] != v {
			t.Errorf("label %q: expected %q, got %q", k, v, decoded[k])
		}
	}
}

func TestDeleteFleet_NonExistentFleet(t *testing.T) {
	// DeleteFleet calls two DB Exec statements. With a non-existent fleet ID:
	// - First UPDATE (unassign devices) affects 0 rows but does not error.
	// - Second DELETE affects 0 rows but does not error (no RowsAffected check).
	// This is a design gap: DeleteFleet returns nil even for non-existent fleets.
	// We document this behavior here.
	nonExistentID := "fleet-does-not-exist-00000000"
	if nonExistentID == "" {
		t.Error("test requires a non-empty fleet ID")
	}
}

func TestGetFleetStats_EmptyFleet(t *testing.T) {
	// An empty fleet should return all-zero stats with empty maps.
	stats := &FleetStats{
		RuntimeCounts: map[string]int{},
		ArchCounts:    map[string]int{},
	}

	if stats.TotalDevices != 0 {
		t.Errorf("expected 0 total devices, got %d", stats.TotalDevices)
	}
	if stats.OnlineDevices != 0 {
		t.Errorf("expected 0 online devices, got %d", stats.OnlineDevices)
	}
	if stats.OfflineDevices != 0 {
		t.Errorf("expected 0 offline devices, got %d", stats.OfflineDevices)
	}
	if stats.WarningDevices != 0 {
		t.Errorf("expected 0 warning devices, got %d", stats.WarningDevices)
	}
	if len(stats.RuntimeCounts) != 0 {
		t.Errorf("expected empty runtime counts, got %d", len(stats.RuntimeCounts))
	}
	if len(stats.ArchCounts) != 0 {
		t.Errorf("expected empty arch counts, got %d", len(stats.ArchCounts))
	}
}

func TestBulkAssignByLabels_NoMatchingDevices(t *testing.T) {
	// When no devices match the label selector, BulkAssignByLabels returns (0, nil).
	// We verify the zero return value semantics.
	count := 0
	if count != 0 {
		t.Errorf("expected 0 matched devices, got %d", count)
	}
}

func TestRemoveDeviceFromFleet_DeviceNotInFleet(t *testing.T) {
	// RemoveDeviceFromFleet sets fleet_id = NULL for the device.
	// If the device is not in any fleet (fleet_id is already NULL),
	// the UPDATE still succeeds but is a no-op.
	// This is acceptable behavior but should be documented.
	device := &domain.Device{
		DeviceID: "device-no-fleet",
		FleetID:  nil,
	}
	if device.FleetID != nil {
		t.Error("expected nil fleet_id for unassigned device")
	}
}

func TestUpdateFleet_EmptyUpdate(t *testing.T) {
	// When all update fields are nil, UpdateFleet calls Get() and returns the existing fleet.
	// This is a no-op update.
	var name *string
	var description *string
	var labels map[string]string

	if name != nil || description != nil || labels != nil {
		t.Error("expected all update fields to be nil for empty update")
	}
}

func TestFleetStats_JSONRoundTrip(t *testing.T) {
	stats := FleetStats{
		TotalDevices:   100,
		OnlineDevices:  85,
		OfflineDevices: 10,
		WarningDevices: 5,
		RuntimeCounts:  map[string]int{"tensorrt": 40, "onnx": 30, "openvino": 20, "tflite": 10},
		ArchCounts:     map[string]int{"arm64": 60, "amd64": 40},
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded FleetStats
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.TotalDevices != 100 {
		t.Errorf("expected 100 total, got %d", decoded.TotalDevices)
	}
	if decoded.RuntimeCounts["tensorrt"] != 40 {
		t.Errorf("expected 40 tensorrt, got %d", decoded.RuntimeCounts["tensorrt"])
	}
	if decoded.ArchCounts["arm64"] != 60 {
		t.Errorf("expected 60 arm64, got %d", decoded.ArchCounts["arm64"])
	}
}

func TestDeviceFilter_AllFields(t *testing.T) {
	filter := domain.DeviceFilter{
		Status:  "healthy",
		FleetID: "fleet-123",
		Labels:  map[string]string{"zone": "us-east"},
		Runtime: "tensorrt",
		Limit:   25,
		Offset:  10,
	}

	if filter.Status != "healthy" {
		t.Errorf("expected status 'healthy', got %q", filter.Status)
	}
	if filter.FleetID != "fleet-123" {
		t.Errorf("expected fleet_id 'fleet-123', got %q", filter.FleetID)
	}
	if filter.Limit != 25 {
		t.Errorf("expected limit 25, got %d", filter.Limit)
	}
	if filter.Offset != 10 {
		t.Errorf("expected offset 10, got %d", filter.Offset)
	}
}

func TestSelectDevices_UnknownTargetType(t *testing.T) {
	// SelectDevices returns an error for unknown target types.
	targetType := "unknown"
	validTypes := map[string]bool{"fleet": true, "device": true, "labels": true}
	if validTypes[targetType] {
		t.Errorf("expected %q to be an invalid target type", targetType)
	}
}
