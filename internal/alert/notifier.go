package alert

import (
	"fmt"
	"log"
	"time"

	"github.com/user/vaultpulse/internal/vault"
)

// Channel represents the type of alert channel.
type Channel string

const (
	ChannelLog   Channel = "log"
	ChannelStdout Channel = "stdout"
)

// Config holds alert configuration.
type Config struct {
	Channel  Channel
	MinLevel string // minimum severity level to alert on: "warning", "critical"
}

// Notifier sends alerts for expiring leases.
type Notifier struct {
	cfg Config
}

// NewNotifier creates a new Notifier with the given config.
func NewNotifier(cfg Config) *Notifier {
	if cfg.Channel == "" {
		cfg.Channel = ChannelStdout
	}
	if cfg.MinLevel == "" {
		cfg.MinLevel = "warning"
	}
	return &Notifier{cfg: cfg}
}

// Notify sends alerts for leases that meet the minimum severity threshold.
func (n *Notifier) Notify(leases []vault.Lease) error {
	for _, l := range leases {
		if !n.shouldAlert(l.Severity) {
			continue
		}
		msg := n.format(l)
		switch n.cfg.Channel {
		case ChannelLog:
			log.Println(msg)
		case ChannelStdout:
			fmt.Println(msg)
		default:
			fmt.Println(msg)
		}
	}
	return nil
}

func (n *Notifier) shouldAlert(severity string) bool {
	rank := map[string]int{
		"ok":       0,
		"warning":  1,
		"critical": 2,
	}
	min := rank[n.cfg.MinLevel]
	cur, ok := rank[severity]
	if !ok {
		return false
	}
	return cur >= min
}

func (n *Notifier) format(l vault.Lease) string {
	ttl := time.Until(l.ExpireTime).Round(time.Second)
	return fmt.Sprintf("[%s] lease %s (path: %s) expires in %s",
		l.Severity, l.LeaseID, l.Path, ttl)
}
