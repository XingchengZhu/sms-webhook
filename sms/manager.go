package sms

import (
    "encoding/json"
    "strings"

    "github.com/sirupsen/logrus"
)

type ProviderConfig struct {
    Name         string `json:"name"`
    Kind         string `json:"kind"` // json | form | header-json
    URL          string `json:"url"`
    Code         string `json:"code"`
    Target       string `json:"target"`

    APIKey    string `json:"api_key"`
    HeaderKey string `json:"header_key"`

    PhoneField   string `json:"phone_field"`
    ContentField string `json:"content_field"`
    CodeField    string `json:"code_field"`
}

type Manager struct {
    senders       map[string]Sender
    defaultTarget string
    sendMode      string
}

func NewManager(jsonStr string, fallback Sender, defaultTarget, sendMode string) *Manager {
    m := &Manager{
        senders:       make(map[string]Sender),
        defaultTarget: defaultTarget,
        sendMode:      sendMode,
    }

    // 1) 解析多通道 JSON
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

    // 2) 完全没配多通道，就用老的单通道
    if len(m.senders) == 0 && fallback != nil {
        m.senders[fallback.Name()] = fallback
    }

    logrus.WithFields(logrus.Fields{
        "senders": m.list(),
        "mode":    m.sendMode,
    }).Info("sms manager initialized")

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
    case "json", "":
        fallthrough
    default:
        return NewJSONSender(c.Name, c.URL, c.Code)
    }
}

func firstNonEmpty(xs ...string) string {
    for _, x := range xs {
        if x != "" {
            return x
        }
    }
    return ""
}

func (m *Manager) list() []string {
    out := make([]string, 0, len(m.senders))
    for name := range m.senders {
        out = append(out, name)
    }
    return out
}

func (m *Manager) SendBroadcast(content, target string) {
    tgt := target
    if tgt == "" {
        tgt = m.defaultTarget
    }
    for name, s := range m.senders {
        if err := s.Send(tgt, content); err != nil {
            logrus.WithError(err).WithField("sender", name).Error("broadcast sms failed")
        }
    }
}

func (m *Manager) SendTo(names []string, content, target string) {
    tgt := target
    if tgt == "" {
        tgt = m.defaultTarget
    }
    for _, name := range names {
        s, ok := m.senders[name]
        if !ok {
            logrus.WithField("sender", name).Warn("sender not found")
            continue
        }
        if err := s.Send(tgt, content); err != nil {
            logrus.WithError(err).WithField("sender", name).Error("send sms failed")
        }
    }
}

// 从告警文本里解析 “渠道: xxx,yyy”
func ParseChannels(alertText string) []string {
    lines := strings.Split(alertText, "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "渠道:") || strings.HasPrefix(line, "channel:") {
            v := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "渠道:"), "channel:"))
            parts := strings.Split(v, ",")
            res := make([]string, 0, len(parts))
            for _, p := range parts {
                p = strings.TrimSpace(p)
                if p != "" {
                    res = append(res, p)
                }
            }
            return res
        }
    }
    return nil
}
