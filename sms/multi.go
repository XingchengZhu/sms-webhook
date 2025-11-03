// sms/multi.go
package sms

import (
	"errors"

	"github.com/sirupsen/logrus"
)

type MultiSender struct {
	Senders []Sender
	Mode    string // "pick" | "broadcast"
}

func (m *MultiSender) Send(target, content string) error {
	if len(m.Senders) == 0 {
		return errors.New("no sender configured")
	}
	switch m.Mode {
	case "pick":
		var lastErr error
		for _, s := range m.Senders {
			if err := s.Send(target, content); err != nil {
				lastErr = err
				logrus.WithError(err).Warn("pick sender failed, trying next")
				continue
			}
			return nil
		}
		return lastErr
	default: // broadcast
		var firstErr error
		for _, s := range m.Senders {
			if err := s.Send(target, content); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	}
}
