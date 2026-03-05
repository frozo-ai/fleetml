# Deployment Policies

FleetML supports policy-driven deployments that control how models are rolled out across device fleets.

## Policy Types

### Deployment Policies

Control deployment behavior and constraints.

```json
{
  "name": "production-safety",
  "policy_type": "deployment",
  "rules": {
    "max_concurrent_deployments": 3,
    "require_canary": true
  },
  "enabled": true,
  "priority": 10,
  "target_fleet_id": "production-fleet-uuid"
}
```

Available rules:
- `max_concurrent_deployments` — Maximum number of simultaneous deployments
- `require_canary` — Force canary deployment policy

### Scaling Policies

Control auto-scaling behavior (future feature).

```json
{
  "policy_type": "scaling",
  "rules": {
    "min_replicas": 2,
    "max_replicas": 10,
    "cpu_threshold": 80
  }
}
```

### Alerting Policies

Define alert conditions and thresholds.

```json
{
  "policy_type": "alerting",
  "rules": {
    "cpu_threshold": 90,
    "memory_threshold": 85,
    "disk_threshold": 95,
    "offline_threshold_minutes": 5
  }
}
```

### Compliance Policies

Enforce organizational requirements.

```json
{
  "policy_type": "compliance",
  "rules": {
    "require_checksum_verification": true,
    "allowed_formats": ["onnx"],
    "max_model_size_mb": 500
  }
}
```

## Canary Deployments

Canary deployments roll out to a small percentage first, then gradually expand:

```
Stage 1: 5% of devices  →  Monitor for 10 minutes
Stage 2: 25% of devices →  Monitor for 30 minutes
Stage 3: 100% of devices
```

### Configuration

When creating a deployment with the `canary` policy:

```json
{
  "model_name": "face-detector",
  "model_version": "v2.0",
  "target_type": "fleet",
  "target_id": "fleet-uuid",
  "policy": "canary",
  "canary_config": {
    "stages": [
      { "percent": 5, "duration": "10m", "success_metric": "error_rate < 0.01" },
      { "percent": 25, "duration": "30m", "success_metric": "error_rate < 0.01" },
      { "percent": 100, "duration": "0", "success_metric": "" }
    ]
  }
}
```

### Automatic Rollback

If a canary stage fails (success metric not met), the deployment automatically:
1. Stops rolling forward
2. Rolls back affected devices to the previous model version
3. Marks the deployment as `rolled_back`

## Policy Evaluation

Policies are evaluated before each deployment:

1. All enabled policies are loaded, ordered by priority (highest first)
2. Fleet-scoped policies only apply to their target fleet
3. If any policy blocks the deployment, it is rejected with a reason
4. Multiple policies can be active simultaneously

## Managing Policies

### CLI
Policies are managed via the REST API or dashboard.

### REST API

```bash
# Create a policy
curl -X POST http://localhost:8080/api/v1/policies \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "max-deploys",
    "policy_type": "deployment",
    "rules": {"max_concurrent_deployments": 5},
    "enabled": true,
    "priority": 10
  }'

# List policies
curl http://localhost:8080/api/v1/policies?type=deployment \
  -H "Authorization: Bearer $TOKEN"

# Disable a policy
curl -X PATCH http://localhost:8080/api/v1/policies/<id> \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"enabled": false}'
```

### Dashboard

Navigate to **Policies** in the sidebar to view, filter, and manage policies. Use the type filter to show only deployment, scaling, alerting, or compliance policies.
