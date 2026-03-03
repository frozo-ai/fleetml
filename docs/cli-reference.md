# FleetML CLI Reference

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--server` | `http://localhost:8080` | Server URL |
| `--api-key` | | API key for authentication |
| `--format` | `table` | Output format (`table`, `json`) |

## Commands

### `fleetml init`

Configure the CLI and test server connectivity.

```bash
fleetml init --server http://your-server:8080
```

Saves configuration to `~/.fleetml/config.yaml`.

### `fleetml version`

Print the CLI version, git commit, and build date.

```bash
fleetml version
```

### `fleetml deploy <model_file>`

Upload a model and deploy it to devices.

```bash
# Basic deployment
fleetml deploy model.onnx --fleet default

# With all options
fleetml deploy model.onnx \
  --name my-model \
  --version 2.0 \
  --fleet production \
  --policy canary \
  --canary-stages "5:5m,50:10m,100:15m" \
  --wait \
  --timeout 30m
```

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | filename | Model name |
| `--version` | `1.0` | Model version |
| `--fleet` | `default` | Target fleet name |
| `--device` | | Target a single device |
| `--labels` | | Label selector (`key=value`) |
| `--policy` | `immediate` | `immediate` or `canary` |
| `--canary-stages` | | Canary config (`percent:duration,...`) |
| `--wait` | `false` | Wait for deployment to complete |
| `--timeout` | `10m` | Wait timeout |

### `fleetml status`

View fleet and deployment status.

```bash
# Fleet overview
fleetml status

# Specific deployment
fleetml status --deployment <deployment-id>
```

| Flag | Default | Description |
|------|---------|-------------|
| `--deployment` | | Show specific deployment |

### `fleetml rollback`

Rollback a deployment to the previous model version.

```bash
# Rollback by deployment ID
fleetml rollback --deployment <deployment-id>

# Rollback a fleet to a specific version
fleetml rollback --fleet production --to-version 1.0
```

| Flag | Default | Description |
|------|---------|-------------|
| `--deployment` | | Deployment ID to rollback |
| `--fleet` | | Fleet to rollback |
| `--device` | | Device to rollback |
| `--to-version` | | Specific version to rollback to |

### `fleetml logs`

View device logs (coming soon).

```bash
fleetml logs --device <device-id> --since 1h --level error
```

| Flag | Default | Description |
|------|---------|-------------|
| `--device` | | Device ID |
| `--follow` | `false` | Follow log output |
| `--since` | `1h` | Show logs since duration |
| `--level` | `info` | Minimum log level |
| `--limit` | `100` | Max log lines |
