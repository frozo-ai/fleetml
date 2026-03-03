package fleet

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fleetml/fleetml/server/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Manager handles fleet and device management.
type Manager struct {
	db *pgxpool.Pool
}

func NewManager(db *pgxpool.Pool) *Manager {
	return &Manager{db: db}
}

// RegisterDevice registers a new device or updates an existing one.
func (m *Manager) RegisterDevice(ctx context.Context, info *domain.Device) (*domain.Device, error) {
	labelsJSON, err := json.Marshal(info.Labels)
	if err != nil {
		return nil, fmt.Errorf("marshal labels: %w", err)
	}

	var d domain.Device
	err = m.db.QueryRow(ctx, `
		INSERT INTO devices (device_id, name, status, arch, gpu_type, runtime, ram_mb, disk_gb, os, hardware_model, labels)
		VALUES ($1, $2, 'registered', $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (device_id) DO UPDATE SET
			arch = EXCLUDED.arch,
			gpu_type = EXCLUDED.gpu_type,
			runtime = EXCLUDED.runtime,
			ram_mb = EXCLUDED.ram_mb,
			disk_gb = EXCLUDED.disk_gb,
			os = EXCLUDED.os,
			hardware_model = EXCLUDED.hardware_model,
			updated_at = NOW()
		RETURNING id, device_id, name, status, arch, gpu_type, runtime, ram_mb, disk_gb, os, hardware_model, registered_at, updated_at`,
		info.DeviceID, info.Name, info.Arch, info.GPUType, info.Runtime,
		info.RAMMB, info.DiskGB, info.OS, info.HardwareModel, labelsJSON,
	).Scan(
		&d.ID, &d.DeviceID, &d.Name, &d.Status, &d.Arch, &d.GPUType,
		&d.Runtime, &d.RAMMB, &d.DiskGB, &d.OS, &d.HardwareModel,
		&d.RegisteredAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("register device: %w", err)
	}

	d.Labels = info.Labels
	return &d, nil
}

// GetDevice returns a device by its device_id.
func (m *Manager) GetDevice(ctx context.Context, deviceID string) (*domain.Device, error) {
	var d domain.Device
	var labelsJSON []byte

	err := m.db.QueryRow(ctx, `
		SELECT id, device_id, name, status, arch, gpu_type, runtime, ram_mb, disk_gb, os,
			   hardware_model, labels, fleet_id, last_heartbeat, registered_at, updated_at,
			   cpu_percent, gpu_percent, ram_mb_used, disk_percent, temperature_c, uptime_hours
		FROM devices WHERE device_id = $1`, deviceID,
	).Scan(
		&d.ID, &d.DeviceID, &d.Name, &d.Status, &d.Arch, &d.GPUType,
		&d.Runtime, &d.RAMMB, &d.DiskGB, &d.OS, &d.HardwareModel,
		&labelsJSON, &d.FleetID, &d.LastHeartbeat, &d.RegisteredAt, &d.UpdatedAt,
		&d.CPUPercent, &d.GPUPercent, &d.RAMMBUsed, &d.DiskPercent,
		&d.TemperatureC, &d.UptimeHours,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("device %s not found", deviceID)
		}
		return nil, fmt.Errorf("get device: %w", err)
	}

	if labelsJSON != nil {
		json.Unmarshal(labelsJSON, &d.Labels)
	}
	return &d, nil
}

// ListDevices lists devices with optional filters.
func (m *Manager) ListDevices(ctx context.Context, filter domain.DeviceFilter) ([]*domain.Device, int, error) {
	query := `SELECT id, device_id, name, status, arch, gpu_type, runtime, ram_mb, disk_gb, os,
			         hardware_model, labels, fleet_id, last_heartbeat, registered_at, updated_at,
			         cpu_percent, gpu_percent, ram_mb_used, disk_percent, temperature_c, uptime_hours
			  FROM devices WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM devices WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.FleetID != "" {
		query += fmt.Sprintf(" AND fleet_id = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND fleet_id = $%d", argIdx)
		args = append(args, filter.FleetID)
		argIdx++
	}
	if filter.Runtime != "" {
		query += fmt.Sprintf(" AND runtime = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND runtime = $%d", argIdx)
		args = append(args, filter.Runtime)
		argIdx++
	}

	var total int
	m.db.QueryRow(ctx, countQuery, args...).Scan(&total)

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	query += fmt.Sprintf(" ORDER BY registered_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := m.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var devices []*domain.Device
	for rows.Next() {
		var d domain.Device
		var labelsJSON []byte
		err := rows.Scan(
			&d.ID, &d.DeviceID, &d.Name, &d.Status, &d.Arch, &d.GPUType,
			&d.Runtime, &d.RAMMB, &d.DiskGB, &d.OS, &d.HardwareModel,
			&labelsJSON, &d.FleetID, &d.LastHeartbeat, &d.RegisteredAt, &d.UpdatedAt,
			&d.CPUPercent, &d.GPUPercent, &d.RAMMBUsed, &d.DiskPercent,
			&d.TemperatureC, &d.UptimeHours,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan device: %w", err)
		}
		if labelsJSON != nil {
			json.Unmarshal(labelsJSON, &d.Labels)
		}
		devices = append(devices, &d)
	}

	return devices, total, nil
}

// UpdateDeviceStatus updates device status and metrics from a heartbeat.
func (m *Manager) UpdateDeviceStatus(ctx context.Context, deviceID string, status string, cpuPct, gpuPct, diskPct, tempC, uptimeH *float64, ramUsed *int) error {
	now := time.Now()
	_, err := m.db.Exec(ctx, `
		UPDATE devices SET
			status = $2,
			last_heartbeat = $3,
			cpu_percent = $4,
			gpu_percent = $5,
			ram_mb_used = $6,
			disk_percent = $7,
			temperature_c = $8,
			uptime_hours = $9,
			updated_at = NOW()
		WHERE device_id = $1`,
		deviceID, status, now, cpuPct, gpuPct, ramUsed, diskPct, tempC, uptimeH,
	)
	if err != nil {
		return fmt.Errorf("update device status: %w", err)
	}
	return nil
}

// CreateFleet creates a new device fleet.
func (m *Manager) CreateFleet(ctx context.Context, name, description string, labels map[string]string) (*domain.Fleet, error) {
	labelsJSON, _ := json.Marshal(labels)

	var f domain.Fleet
	err := m.db.QueryRow(ctx, `
		INSERT INTO fleets (name, description, labels)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, created_at, updated_at`,
		name, description, labelsJSON,
	).Scan(&f.ID, &f.Name, &f.Description, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create fleet: %w", err)
	}

	f.Labels = labels
	return &f, nil
}

// AssignDeviceToFleet adds a device to a fleet.
func (m *Manager) AssignDeviceToFleet(ctx context.Context, deviceID string, fleetID string) error {
	_, err := m.db.Exec(ctx, `
		UPDATE devices SET fleet_id = $2, updated_at = NOW()
		WHERE device_id = $1`, deviceID, fleetID,
	)
	if err != nil {
		return fmt.Errorf("assign device to fleet: %w", err)
	}
	return nil
}

// SelectDevices selects devices matching a deployment target.
func (m *Manager) SelectDevices(ctx context.Context, targetType, targetID string, targetLabels map[string]string) ([]*domain.Device, error) {
	var query string
	var args []interface{}

	switch targetType {
	case "fleet":
		query = `SELECT id, device_id, name, status, arch, gpu_type, runtime, ram_mb, disk_gb, os, hardware_model, labels
				 FROM devices WHERE fleet_id = $1 AND status != 'decommissioned'`
		args = []interface{}{targetID}
	case "device":
		query = `SELECT id, device_id, name, status, arch, gpu_type, runtime, ram_mb, disk_gb, os, hardware_model, labels
				 FROM devices WHERE device_id = $1`
		args = []interface{}{targetID}
	case "labels":
		labelsJSON, _ := json.Marshal(targetLabels)
		query = `SELECT id, device_id, name, status, arch, gpu_type, runtime, ram_mb, disk_gb, os, hardware_model, labels
				 FROM devices WHERE labels @> $1 AND status != 'decommissioned'`
		args = []interface{}{labelsJSON}
	default:
		return nil, fmt.Errorf("unknown target type: %s", targetType)
	}

	rows, err := m.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select devices: %w", err)
	}
	defer rows.Close()

	var devices []*domain.Device
	for rows.Next() {
		var d domain.Device
		var labelsJSON []byte
		err := rows.Scan(
			&d.ID, &d.DeviceID, &d.Name, &d.Status, &d.Arch, &d.GPUType,
			&d.Runtime, &d.RAMMB, &d.DiskGB, &d.OS, &d.HardwareModel, &labelsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		if labelsJSON != nil {
			json.Unmarshal(labelsJSON, &d.Labels)
		}
		devices = append(devices, &d)
	}

	return devices, nil
}

// ListFleets lists all fleets.
func (m *Manager) ListFleets(ctx context.Context) ([]*domain.Fleet, error) {
	rows, err := m.db.Query(ctx, `
		SELECT id, name, description, labels, created_at, updated_at
		FROM fleets ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list fleets: %w", err)
	}
	defer rows.Close()

	var fleets []*domain.Fleet
	for rows.Next() {
		var f domain.Fleet
		var labelsJSON []byte
		err := rows.Scan(&f.ID, &f.Name, &f.Description, &labelsJSON, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan fleet: %w", err)
		}
		if labelsJSON != nil {
			json.Unmarshal(labelsJSON, &f.Labels)
		}
		fleets = append(fleets, &f)
	}

	return fleets, nil
}

// UpdateFleet updates a fleet's name, description, and/or labels.
func (m *Manager) UpdateFleet(ctx context.Context, id string, name *string, description *string, labels map[string]string) (*domain.Fleet, error) {
	// Build dynamic update
	if name != nil {
		_, err := m.db.Exec(ctx, `UPDATE fleets SET name = $2, updated_at = NOW() WHERE id = $1`, id, *name)
		if err != nil {
			return nil, fmt.Errorf("update fleet name: %w", err)
		}
	}
	if description != nil {
		_, err := m.db.Exec(ctx, `UPDATE fleets SET description = $2, updated_at = NOW() WHERE id = $1`, id, *description)
		if err != nil {
			return nil, fmt.Errorf("update fleet description: %w", err)
		}
	}
	if labels != nil {
		labelsJSON, _ := json.Marshal(labels)
		_, err := m.db.Exec(ctx, `UPDATE fleets SET labels = $2, updated_at = NOW() WHERE id = $1`, id, labelsJSON)
		if err != nil {
			return nil, fmt.Errorf("update fleet labels: %w", err)
		}
	}

	return m.GetFleet(ctx, id)
}

// DeleteFleet deletes a fleet and unassigns all devices from it.
func (m *Manager) DeleteFleet(ctx context.Context, id string) error {
	// Unassign devices from fleet first
	_, err := m.db.Exec(ctx, `UPDATE devices SET fleet_id = NULL, updated_at = NOW() WHERE fleet_id = $1`, id)
	if err != nil {
		return fmt.Errorf("unassign fleet devices: %w", err)
	}

	_, err = m.db.Exec(ctx, `DELETE FROM fleets WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete fleet: %w", err)
	}
	return nil
}

// GetFleet returns a fleet by ID.
func (m *Manager) GetFleet(ctx context.Context, id string) (*domain.Fleet, error) {
	var f domain.Fleet
	var labelsJSON []byte

	err := m.db.QueryRow(ctx, `
		SELECT id, name, description, labels, created_at, updated_at
		FROM fleets WHERE id = $1`, id,
	).Scan(&f.ID, &f.Name, &f.Description, &labelsJSON, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("fleet %s not found", id)
		}
		return nil, fmt.Errorf("get fleet: %w", err)
	}

	if labelsJSON != nil {
		json.Unmarshal(labelsJSON, &f.Labels)
	}
	return &f, nil
}
