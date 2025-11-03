package sms

import (
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"
)

type ProviderConfig struct {
	Name         string `json:"name"`
	Kind         string `json:"kind"` // json | form | header-json | feishu
	URL          string `json:"url"`
	Code         string `json:"code"`
	Target       string `json:"target"`
	APIKey       string `json:"api_key"`   // 兼容旧字段
	HeaderKey    string `json:"header_key"`
	PhoneField   string `json:"phone_field"`
	ContentField string `json:"content_field"`
	CodeField    string `json:"code_field"`
	Secret       string `json:"secret"` // 新增：飞书签名用
}

// Manager 管一堆 sender
type Manager struct {
	senders       map[string]Sender
	defaultTarget string
	sendMode      string // pick | broadcast
	primary       string // pick 模式下用谁
}

// jsonStr: 来自 SMS_PROVIDERS_JSON
// fallback: 老的单通道 sender
func NewManager(jsonStr string, fallback Sender, defaultTarget, sendMode string) *Manager {
	m := &Manager{
		senders:       make(map[string]Sender),
		defaultTarget: defaultTarget,
		sendMode:      sendMode,
	}

	// 1) 多通道 JSON
	if jsonStr != "" {
		var cfgs []ProviderConfig
		if err := json.Unmarshal([]byte(jsonStr), &cfgs); err != nil {
			logrus.WithError(err).Error("parse SMS_PROVIDERS_JSON failed")
		} else {
			for _, c := range cfgs {
				s := buildSenderFromConfig(c)
				if s != nil {
					m.senders[s.Name()] = s
				}
			}
		}
	}

	// 2) 一个都没配，就用老的单通道兜底
	if len(m.senders) == 0 && fallback != nil {
		m.senders[fallback.Name()] = fallback
	}

	// 3) 选 primary
	m.primary = ""
	if len(m.senders) > 0 {
		if _, ok := m.senders["default"]; ok {
			m.primary = "default"
		} else {
			for name := range m.senders {
				m.primary = name
				break
			}
		}
	}

	if m.sendMode == "" {
		m.sendMode = "pick" // 默认只发一条
	}

	logrus.WithFields(logrus.Fields{
		"senders": m.list(),
		"mode":    m.sendMode,
		"primary": m.primary,
	}).Info("sms manager inited")

	return m
}

func buildSenderFromConfig(c ProviderConfig) Sender {
	switch c.Kind {
	case "form":
		return NewFormSender(
			c.Name,
			c.URL,
			firstNonEmpty(c.CodeField, "code"),
			firstNonEmpty(c.PhoneField, "target"),
			firstNonEmpty(c.ContentField, "content"),
			c.Code,
		)
	case "header-json":
		return NewHeaderJSONSender(
			c.Name,
			c.URL,
			c.Code,
			c.APIKey,
			firstNonEmpty(c.HeaderKey, "X-API-KEY"),
		)
	case "feishu":
		// 飞书签名 secret 可用新字段 secret；兼容 api_key
		secret := firstNonEmpty(c.Secret, c.APIKey)
		return NewFeishuSender(
			firstNonEmpty(c.Name, "feishu"),
			c.URL,
			secret,
		)
	case "json", "":
		fallthrough
	default:
		return NewJSONSender(c.Name, c.URL, c.Code)
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func (m *Manager) list() []string {
	out := make([]string, 0, len(m.senders))
	for k := range m.senders {
		out = append(out, k)
	}
	return out
}

// 广播所有
func (m *Manager) SendBroadcast(content, target string) {
	tgt := target
	if tgt == "" {
		tgt = m.defaultTarget
	}
	for name, s := range m.senders {
		if err := s.Send(tgt, content); err != nil {
			logrus.WithError(err).WithField("sender", name).Error("broadcast failed")
		}
	}
}

// 给 handler 用的“默认发送”
// pick: 发 primary
// broadcast: 全发
func (m *Manager) SendDefault(content, target string) {
	if m.sendMode == "broadcast" {
		m.SendBroadcast(content, target)
		return
	}
	// pick 模式
	if m.primary == "" {
		logrus.Warn("no primary sender to send default sms")
		return
	}
	m.SendTo([]string{m.primary}, content, target)
}

func (m *Manager) SendTo(names []string, content, target string) {
	tgt := target
	if tgt == "" {
		tgt = m.defaultTarget
	}

	sentOK := false

	for _, name := range names {
		s, ok := m.senders[name]
		if !ok {
			logrus.WithField("sender", name).Warn("sender not found")
			continue
		}
		if err := s.Send(tgt, content); err != nil {
			logrus.WithError(err).WithField("sender", name).Error("send failed")
			continue
		}
		// 成功一次
		sentOK = true
		if m.sendMode == "pick" {
			// 首个成功就停
			return
		}
	}

	if !sentOK && m.sendMode == "pick" {
		logrus.Warn("no sender succeeded in pick mode")
	}
}

// ParseChannels 从 body 里找“渠道: x,y”
func ParseChannels(alertText string) []string {
	lines := strings.Split(alertText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "渠道:") || strings.HasPrefix(line, "channel:") {
			v := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "渠道:"), "channel:"))
			parts := strings.Split(v, ",")
			out := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					out = append(out, p)
				}
			}
			return out
		}
	}
	return nil
}
