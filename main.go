// main.go（只展示 buildSender 的更新部分）
package main

import (
	"net/http"

	"sms-webhook/config"
	"sms-webhook/handlers"
	"sms-webhook/sms"

	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.LoadConfig()
	logrus.SetLevel(cfg.LogLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	sender := buildSender(cfg)

	http.HandleFunc("/webhook", handlers.WebhookHandler(cfg, sender))
	logrus.WithFields(logrus.Fields{
		"port":     cfg.Port,
		"provider": cfg.SMSProvider,
	}).Info("Starting webhook server")
	logrus.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}

func buildSender(cfg config.Config) sms.Sender {
	// 组合策略
	switch cfg.SMSProvider {
	case "broadcast":
		return &sms.MultiSender{Senders: buildMany(cfg, cfg.SMSProviders), Mode: "broadcast"}
	case "pick":
		return &sms.MultiSender{Senders: buildMany(cfg, cfg.SMSProviders), Mode: "pick"}
	}

	// 单路
	return buildOne(cfg, cfg.SMSProvider)
}

func buildMany(cfg config.Config, names []string) []sms.Sender {
	out := make([]sms.Sender, 0, len(names))
	for _, n := range names {
		if s := buildOne(cfg, n); s != nil {
			out = append(out, s)
		}
	}
	return out
}

func buildOne(cfg config.Config, name string) sms.Sender {
	switch name {
	case "form":
		return &sms.FormSender{
			URL:          cfg.SMSAPIURL,
			CodeField:    "code",
			PhoneField:   "target",
			ContentField: "content",
			CodeValue:    cfg.SMSCode,
		}
	case "header-json":
		return &sms.HeaderJSONSender{
			URL:       cfg.SMSAPIURL,
			Code:      cfg.SMSCode,
			APIKey:    cfg.SMSAPIKey,
			HeaderKey: cfg.SMSHeaderKey,
		}
	case "feishu-webhook":
		return &sms.FeishuWebhookSender{
			WebhookURL: cfg.FeishuWebhook,
			Secret:     cfg.FeishuSecret,
		}
	case "feishu-api":
		return &sms.FeishuAPISender{
			AppID:          cfg.FeishuAppID,
			AppSecret:      cfg.FeishuAppSecret,
			ReceiveIDType:  cfg.FeishuReceiveIDType,
			ReceiveID:      cfg.FeishuReceiveID,
		}
	default: // "json"
		return &sms.JSONSender{
			URL:  cfg.SMSAPIURL,
			Code: cfg.SMSCode,
		}
	}
}
