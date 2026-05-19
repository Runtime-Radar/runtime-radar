package syslog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	syslog "github.com/hashicorp/go-syslog"
	"github.com/runtime-radar/runtime-radar/notifier/api"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/model"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/template"
	enforcer "github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/model"
)

var (
	TestConnectionTimeout = 10 * time.Second
)

type Notifier struct {
	Config *model.Syslog
}

func (n *Notifier) Test(_ context.Context) error {
	u, err := url.Parse(n.Config.Address)
	if err != nil {
		return fmt.Errorf("can't prase address: %w", err)
	}

	switch u.Scheme {
	case "udp":
		// no check for udp
	case "tcp":
		conn, err := net.DialTimeout(u.Scheme, u.Host, TestConnectionTimeout)
		if err != nil {
			return fmt.Errorf("can't connect to address: %w", err)
		}
		defer conn.Close()
	default:
		return fmt.Errorf("unknown protocol: %s", u.Scheme)
	}

	return nil
}

func (n *Notifier) Notify(_ context.Context, notification *model.Notification, event any) error {
	data, err := renderJSON(notification, event)
	if err != nil {
		return fmt.Errorf("can't render json: %w", err)
	}

	u, err := url.Parse(n.Config.Address)
	if err != nil {
		return fmt.Errorf("can't parse address '%s': %w", n.Config.Address, err)
	}

	priority := mapEventPriorityToSyslog(event)

	w, err := syslog.DialLogger(u.Scheme, u.Host, priority, "MAIL", "CS")
	if err != nil {
		return fmt.Errorf("can't connect to syslog: %w", err)
	}
	defer w.Close()

	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("can't write message to syslog: %w", err)
	}

	return nil
}

func renderJSON(n *model.Notification, event any) ([]byte, error) {
	var unwrappedEvent any

	switch ev := event.(type) {
	case *api.Message_RuntimeEvent:
		unwrappedEvent = ev.RuntimeEvent
	default:
		return nil, fmt.Errorf("invalid event type given: %T", ev)
	}

	tplData := map[string]any{
		"event":            unwrappedEvent,
		"notificationName": n.Name,
		"centralCSURL":     n.CentralCSURL,
		"csClusterID":      n.CSClusterID,
		"csClusterName":    n.CSClusterName,
		"ownCSURL":         n.OwnCSURL,
	}

	tpl, ok := template.DefaultTexts[n.EventType]
	if !ok {
		return nil, fmt.Errorf("no default template for '%s' event type", n.EventType)
	}

	if n.Template != "" {
		var err error

		tpl, err = template.NewText("", n.Template)
		if err != nil {
			return nil, fmt.Errorf("can't parse template: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, tplData); err != nil {
		return nil, fmt.Errorf("can't execute template: %w", err)
	}

	var compactBuf bytes.Buffer
	if err := json.Compact(&compactBuf, buf.Bytes()); err != nil {
		return nil, fmt.Errorf("can't compact json: %w", err)
	}

	return compactBuf.Bytes(), nil
}

func mapEventPriorityToSyslog(event any) syslog.Priority {
	switch ev := event.(type) {
	case *api.Message_RuntimeEvent:
		return mapSeverityToSyslog(ev.RuntimeEvent.Severity)
	}
	return syslog.LOG_DEBUG
}

func mapSeverityToSyslog(severity string) syslog.Priority {
	switch strings.ToLower(severity) {
	case enforcer.CriticalSeverity.String():
		return syslog.LOG_EMERG
	case enforcer.HighSeverity.String():
		return syslog.LOG_ALERT
	case enforcer.MediumSeverity.String():
		return syslog.LOG_CRIT
	case enforcer.LowSeverity.String():
		return syslog.LOG_WARNING
	case "info":
		return syslog.LOG_INFO
	case "unknown":
		return syslog.LOG_NOTICE
	case "":
		return syslog.LOG_ERR
	}
	return syslog.LOG_DEBUG
}
