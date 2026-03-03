# FleetML REST API Reference

Base URL: `http://localhost:8080/api/v1`

All endpoints require `Authorization: Bearer <token>` unless noted.

## Health

### `GET /health`
Health check (no auth required).

**Response:** `200 OK`
```json
{"status": "ok", "version": "0.1.0"}
```

## Authentication

### `POST /api/v1/auth/register`
Register a new user (no auth required on first setup).

**Request:**
```json
{
  "email": "admin@example.com",
  "password": "secure-password",
  "name": "Admin User"
}
```

### `POST /api/v1/auth/login`
Authenticate and receive JWT tokens.

**Request:**
```json
{
  "email": "admin@example.com",
  "password": "secure-password"
}
```

**Response:**
```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_at": "2026-03-02T12:00:00Z",
  "token_type": "Bearer"
}
```

### `GET /api/v1/auth/me`
Get current user info. **Requires:** any role.

## Models

### `POST /api/v1/models`
Upload a model (multipart form). **Requires:** `models:write`.

**Form fields:**
- `file` — Model file (max 500MB)
- `name` — Model name
- `version` — Model version
- `format` — `onnx`, `tensorrt`, `openvino`, `tflite`

### `GET /api/v1/models`
List models. **Requires:** `models:read`.

**Query params:** `name`, `tags`, `limit`, `offset`

### `GET /api/v1/models/{name}/{version}`
Get a specific model. **Requires:** `models:read`.

### `DELETE /api/v1/models/{id}`
Delete a model. **Requires:** `models:delete`.

## Devices

### `GET /api/v1/devices`
List devices. **Requires:** `devices:read`.

**Query params:** `status`, `fleet_id`, `runtime`, `limit`, `offset`

### `GET /api/v1/devices/{device_id}`
Get device details. **Requires:** `devices:read`.

### `PUT /api/v1/devices/{device_id}`
Update device (e.g., assign to fleet). **Requires:** `devices:write`.

### `DELETE /api/v1/devices/{device_id}`
Decommission a device. **Requires:** `devices:delete`.

### `GET /api/v1/devices/{device_id}/metrics?since=1h`
Get device metrics history. **Requires:** `devices:read`.

## Fleets

### `POST /api/v1/fleets`
Create a fleet. **Requires:** `fleets:write`.

**Request:**
```json
{
  "name": "production",
  "description": "Production edge devices",
  "labels": {"env": "prod"}
}
```

### `GET /api/v1/fleets`
List fleets. **Requires:** `fleets:read`.

### `GET /api/v1/fleets/{id}`
Get fleet details. **Requires:** `fleets:read`.

## Deployments

### `POST /api/v1/deployments`
Create a deployment. **Requires:** `deployments:write`.

**Request:**
```json
{
  "model_name": "my-model",
  "model_version": "1.0",
  "target_type": "fleet",
  "target_id": "fleet-uuid",
  "policy": "canary",
  "canary_config": {
    "stages": [
      {"percent": 5, "duration": "5m", "success_metric": "error_rate < 1%"},
      {"percent": 50, "duration": "10m"},
      {"percent": 100, "duration": "15m"}
    ]
  }
}
```

### `GET /api/v1/deployments`
List deployments. **Requires:** `deployments:read`.

**Query params:** `state`, `model_name`

### `GET /api/v1/deployments/{id}`
Get deployment with per-device status. **Requires:** `deployments:read`.

### `POST /api/v1/deployments/{id}/cancel`
Cancel a deployment. **Requires:** `deployments:cancel`.

### `POST /api/v1/deployments/{id}/rollback`
Rollback a deployment. **Requires:** `deployments:write`.

## Fleet Summary

### `GET /api/v1/health/fleet-summary`
Get fleet-wide health summary. **Requires:** `devices:read`.

**Response:**
```json
{
  "total_devices": 100,
  "healthy": 85,
  "warning": 10,
  "offline": 5,
  "active_deployments": 2
}
```
