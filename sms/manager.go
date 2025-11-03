// sms/manager.go
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

    // for header-json
    APIKey    string `json:"api_key"`
    HeaderKey string `json:"header_key"`

    // for form
    PhoneField   string `json:"phone_field"`
    ContentField string `json:"content_field"`
    CodeField    string `json:"code_field"`
}

type Manager struct {
    senders   map[string]Sender
    defaultTarget string
    sendMode  string // broadcast | pick
}

func NewManager(jsonStr string, fallback Sender, defaultTarget, sendMode string) *Manager {
    m := &Manager{
        senders:       make(map[string]Sender),
        defaultTarget: defaultTarget,
        sendMode:      sendMode,
    }

    // 1) 如果有 json 配置，先解析
    if jsonStr != "" {
        var cfgs []ProviderConfig
        if err := json.Unmarshal([]byte(jsonStr), &cfgs); err != nil {
            logrus.WithError(err).Error("failed to parse SMS_PROVIDERS_JSON")
        } else {
            for _, c := range cfgs {
                s := buildSenderFromConfig(c)
                if s != nil {
                    m.senders[s.Name()] = s
                }
            }
        }
    }

    // 2) 如果啥都没有，又给了 fallback，就加进去
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
    case "json", "":
        return NewJSONSender(c.Name, c.URL, c.Code)
    case "form":
        return NewFormSender(c) // 你要写一个 NewFormSender
    case "header-json":
        return NewHeaderJSONSender(c) // 你要写一个 NewHeaderJSONSender
    default:
        logrus.WithField("kind", c.Kind).Warn("unknown sms kind")
        return nil
    }
}

// list just for log
func (m *Manager) list() []string {
    names := make([]string, 0, len(m.senders))
    for name := range m.senders {
        names = append(names, name)
    }
    return names
}

// SendBroadcast: 发给所有人
func (m *Manager) SendBroadcast(content string, targetOverride string) {
    target := targetOverride
    if target == "" {
        target = m.defaultTarget
    }
    for name, s := range m.senders {
        if err := s.Send(target, content); err != nil {
            logrus.WithError(err).WithField("sender", name).Error("broadcast sms failed")
        }
    }
}

// SendTo: 发给指定的那几个
func (m *Manager) SendTo(names []string, content string, targetOverride string) {
    target := targetOverride
    if target == "" {
        target = m.defaultTarget
    }
    for _, name := range names {
        s, ok := m.senders[name]
        if !ok {
            logrus.WithField("sender", name).Warn("sender not found")
            continue
        }
        if err := s.Send(target, content); err != nil {
            logrus.WithError(err).WithField("sender", name).Error("send sms failed")
        }
    }
}

// ParseProvidersFromAlert: 从告警文本里解析你想要的通道
// 例如某一行是：渠道: json1,header1
func ParseProvidersFromAlert(alertText string) []string {
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
