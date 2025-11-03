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

        parts := strings.Split(raw, "\n\n")
        for _, p := range parts {
            p = strings.TrimSpace(p)
            if p == "" {
                continue
            }

            // 1. 取描述
            summary := extractSummary(p)
            if summary == "" {
                summary = "No summary provided"
            }

            // 2. 看有没有写渠道
            chs := sms.ParseChannels(p)
            if len(chs) == 0 {
                manager.SendBroadcast(summary, "")
            } else {
                manager.SendTo(chs, summary, "")
            }

            logrus.WithField("content", summary).Info("SMS processed")
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
