# vaultpulse

A CLI tool that monitors HashiCorp Vault secret lease expirations and sends configurable alerts before they expire.

---

## Installation

```bash
go install github.com/yourusername/vaultpulse@latest
```

Or download a pre-built binary from the [releases page](https://github.com/yourusername/vaultpulse/releases).

---

## Usage

Set your Vault address and token, then run vaultpulse with a config file:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.xxxxxxxxxxxxxxxx"

vaultpulse --config config.yaml
```

**Example `config.yaml`:**

```yaml
alert_threshold: 72h
check_interval: 30m
notifiers:
  - type: slack
    webhook_url: "https://hooks.slack.com/services/..."
  - type: email
    recipients:
      - ops@example.com
paths:
  - secret/prod/*
  - database/creds/my-role
```

vaultpulse will poll Vault at the configured interval and send alerts when any monitored lease is within the specified threshold of expiration.

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to config file | `config.yaml` |
| `--log-level` | Log verbosity (`debug`, `info`, `warn`) | `info` |
| `--dry-run` | Check leases without sending alerts | `false` |

---

## License

MIT © [yourusername](https://github.com/yourusername)