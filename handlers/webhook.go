package handlers

import (
	"io"
	"net/http"
	"sms-webhook/config"
	"sms-webhook/utils"
	"strings"

	"github.com/sirupsen/logrus"
)

func WebhookHandler(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logrus.WithError(err).Error("Failed to read request body")
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		logrus.WithField("body", string(body)).Debug("Received raw request body")

		// 简单的字符串解析器，根据预期格式提取信息
		alerts := strings.Split(string(body), "\n\n") // 假设不同的告警之间用两个换行符分隔

		for _, alertText := range alerts {
			alertLines := strings.Split(alertText, "\n")
			summary := ""
			for _, line := range alertLines {
				if strings.HasPrefix(line, "描述: ") {
					summary = strings.TrimPrefix(line, "描述: ")
					break
				}
			}

			if summary == "" {
				summary = "No summary provided"
			}

			sms := utils.SMSRequest{
				Code:    cfg.SMSCode,
				Target:  cfg.SMSTarget,
				Content: summary,
			}

			logrus.WithFields(logrus.Fields{
				"content": sms.Content,
			}).Info("Sending SMS")

			err := utils.SendSMS(cfg, sms)
			if err != nil {
				logrus.WithError(err).Error("Failed to send SMS")
				http.Error(w, "Failed to send SMS", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Alert received and SMS sent"))
	}
}
