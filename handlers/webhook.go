package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/XingchengZhu/sms-webhook/config"
	"github.com/XingchengZhu/sms-webhook/sms"
)

// /webhook 的入口
func WebhookHandler(cfg config.Config, mgr *sms.Manager) http.HandlerFunc {
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

		// 支持一次发多段：用空行分开
		chunks := strings.Split(raw, "\n\n")
		for _, chunk := range chunks {
			chunk = strings.TrimSpace(chunk)
			if chunk == "" {
				continue
			}

			// 1. 取描述
			desc := parseDescription(chunk)
			if desc == "" {
				desc = "No summary provided"
			}

			// 2. 看有没有写渠道
			channels := sms.ParseChannels(chunk)

			if len(channels) > 0 {
				// 显式点名渠道
				mgr.SendTo(channels, desc, cfg.SMSTarget)
			} else {
				// 没点名 → 用当前模式（默认 pick，只发一条）
				mgr.SendDefault(desc, cfg.SMSTarget)
			}

			logrus.WithField("content", desc).Info("SMS processed")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Alert received and SMS sent"))
	}
}

func parseDescription(s string) string {
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "描述:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "描述:"))
		}
	}
	return ""
}
