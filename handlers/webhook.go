package handlers

import (
    "io"
    "net/http"
    "strings"

    "sms-webhook/config"
    "sms-webhook/sms"

    "github.com/sirupsen/logrus"
)

func WebhookHandler(cfg config.Config, manager *sms.Manager) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }

        b, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "read body error", http.StatusInternalServerError)
            return
        }
        defer r.Body.Close()

        raw := string(b)
        logrus.WithField("body", raw).Debug("received webhook")

        // 一次请求里可能有多条告警，用空行分
        parts := strings.Split(raw, "\n\n")
        for _, part := range parts {
            part = strings.TrimSpace(part)
            if part == "" {
                continue
            }

            // 1. 抽内容
            summary := extractSummary(part)
            if summary == "" {
                summary = "No summary provided"
            }

            // 2. 看有没有渠道
            channels := sms.ParseChannels(part)
            if len(channels) == 0 {
                manager.SendBroadcast(summary, "")
            } else {
                manager.SendTo(channels, summary, "")
            }

            logrus.WithField("content", summary).Info("alert processed")
        }

        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Alert received and SMS sent"))
    }
}

func extractSummary(s string) string {
    lines := strings.Split(s, "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "描述: ") {
            return strings.TrimSpace(strings.TrimPrefix(line, "描述: "))
        }
    }
    return ""
}
