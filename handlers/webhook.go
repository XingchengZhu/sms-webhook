// handlers/webhook.go（覆盖）
package handlers

import (
	"encoding/json"
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

		logrus.WithFields(logrus.Fields{
			"ct":   r.Header.Get("Content-Type"),
			"body": string(body),
		}).Debug("received webhook")

		ct := r.Header.Get("Content-Type")
		var msgs []string

		if strings.HasPrefix(ct, "application/json") {
			// 兼容 Alertmanager JSON
			if m := extractFromAlertmanagerJSON(body); len(m) > 0 {
				msgs = m
			}
		}

		if len(msgs) == 0 {
			// 回退到“查找 描述: ”模式（与你现在一致）
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
				msgs = append(msgs, summary)
			}
		}

		for _, m := range msgs {
			if err := sender.Send(cfg.SMSTarget, m); err != nil {
				logrus.WithError(err).Error("send failed")
				http.Error(w, "Failed to send", http.StatusBadGateway)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}

func extractFromAlertmanagerJSON(body []byte) []string {
	var p struct {
		Status string `json:"status"`
		Alerts []struct {
			Status      string            `json:"status"`
			Labels      map[string]string `json:"labels"`
			Annotations map[string]string `json:"annotations"`
		} `json:"alerts"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return nil
	}
	var out []string
	for _, a := range p.Alerts {
		s := a.Annotations["summary"]
		if s == "" {
			s = a.Annotations["description"]
		}
		if s == "" {
			// 兜底拼装
			s = "[" + a.Labels["alertname"] + "] " + a.Labels["severity"] + " " + a.Labels["instance"]
		}
		out = append(out, s)
	}
	return out
}
