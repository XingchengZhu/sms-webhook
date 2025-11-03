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

func WebhookHandler(cfg config.Config, sender sms.Sender) http.HandlerFunc {
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

        logrus.WithField("body", string(body)).Debug("received webhook")

        // 还是按你之前的“描述: ”来
        alerts := strings.Split(string(body), "\n\n")
        for _, alertText := range alerts {
            lines := strings.Split(alertText, "\n")
            summary := ""
            for _, line := range lines {
                if strings.HasPrefix(line, "描述: ") {
                    summary = strings.TrimPrefix(line, "描述: ")
                    break
                }
            }
            if summary == "" {
                summary = "No summary provided"
            }

            // 这里不用关心是哪种短信接口了
            if err := sender.Send(cfg.SMSTarget, summary); err != nil {
                logrus.WithError(err).Error("send sms failed")
                http.Error(w, "Failed to send SMS", http.StatusBadGateway)
                return
            }
        }

        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    }
}
