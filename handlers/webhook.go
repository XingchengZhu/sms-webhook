package handlers

import (
    "encoding/json"
    "io"
    "net/http"
    "strings"

    "github.com/sirupsen/logrus"

    "github.com/XingchengZhu/sms-webhook/config"
    "github.com/XingchengZhu/sms-webhook/sms"
)

// 如果 Content-Type 是 application/json，我们尝试按 alertmanager 的格式取描述
type alertmanagerPayload struct {
    Alerts []struct {
        Annotations map[string]string `json:"annotations"`
    } `json:"alerts"`
}

type WebhookHandler struct {
    sender *sms.Sender
    cfg    config.Config
    log    *logrus.Logger
}

func NewWebhookHandler(sender *sms.Sender, cfg config.Config, log *logrus.Logger) *WebhookHandler {
    return &WebhookHandler{
        sender: sender,
        cfg:    cfg,
        log:    log,
    }
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    bodyBytes, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "failed to read body", http.StatusBadRequest)
        return
    }
    body := strings.TrimSpace(string(bodyBytes))
    h.log.WithField("body", body).Debug("received webhook")

    var channels []string
    var content string

    // 1) JSON 格式（比如 alertmanager）
    if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
        var p alertmanagerPayload
        if err := json.Unmarshal(bodyBytes, &p); err == nil && len(p.Alerts) > 0 {
            // 很简单地拿第一条
            if msg := p.Alerts[0].Annotations["description"]; msg != "" {
                content = msg
            } else if msg := p.Alerts[0].Annotations["summary"]; msg != "" {
                content = msg
            }
        }
    }

    // 2) 文本格式（你现在就是这个）
    if content == "" {
        channels, content = extractChannelsAndContent(body)
    }

    // 3) 决定要发到哪些通道
    targetChannels := channels
    if len(targetChannels) == 0 {
        if len(h.cfg.DefaultChannels) > 0 {
            targetChannels = h.cfg.DefaultChannels
        } else {
            // 没配默认，就全部广播
            for name := range h.cfg.Channels {
                targetChannels = append(targetChannels, name)
            }
        }
    }

    errs := h.sender.SendMany(targetChannels, content)
    if len(errs) > 0 {
        h.log.WithField("errors", errs).Error("some channels failed")
        http.Error(w, "failed to send to some channels", http.StatusInternalServerError)
        return
    }

    w.Write([]byte("Alert received and SMS sent"))
}

func extractChannelsAndContent(body string) ([]string, string) {
    lines := strings.Split(body, "\n")
    var channels []string
    var contentLines []string

    for _, line := range lines {
        trim := strings.TrimSpace(line)
        lower := strings.ToLower(trim)

        // 渠道: json1,header1
        if strings.HasPrefix(trim, "渠道:") || strings.HasPrefix(lower, "channels:") || strings.HasPrefix(lower, "channel:") {
            idx := strings.Index(trim, ":")
            if idx > -1 {
                arr := strings.Split(trim[idx+1:], ",")
                for _, a := range arr {
                    a = strings.TrimSpace(a)
                    if a != "" {
                        channels = append(channels, a)
                    }
                }
            }
            continue
        }

        // 描述: xxx
        if strings.HasPrefix(trim, "描述:") || strings.HasPrefix(lower, "desc:") || strings.HasPrefix(lower, "content:") {
            idx := strings.Index(trim, ":")
            if idx > -1 {
                contentLines = append(contentLines, strings.TrimSpace(trim[idx+1:]))
                continue
            }
        }

        // 其他行也当正文
        if trim != "" {
            contentLines = append(contentLines, trim)
        }
    }

    return channels, strings.Join(contentLines, "\n")
}
