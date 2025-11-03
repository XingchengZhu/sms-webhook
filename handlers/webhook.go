// handlers/webhook.go
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

        body, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "read body error", http.StatusInternalServerError)
            return
        }
        defer r.Body.Close()

        raw := string(body)
        logrus.WithField("body", raw).Debug("received webhook")

        // 可能是多条告警
        chunks := strings.Split(raw, "\n\n")

        for _, chunk := range chunks {
            if strings.TrimSpace(chunk) == "" {
                continue
            }

            // 1) 拿短信内容
            summary := extractSummary(chunk)
            if summary == "" {
                summary = "No summary provided"
            }

            // 2) 看有没有指定渠道
            names := sms.ParseProvidersFromAlert(chunk)

            if len(names) == 0 {
                // 没指定，按配置发
                manager.SendBroadcast(summary, "")
            } else {
                manager.SendTo(names, summary, "")
            }
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
