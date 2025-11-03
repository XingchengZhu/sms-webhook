package main

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/XingchengZhu/sms-webhook/config"
	"github.com/XingchengZhu/sms-webhook/handlers"
	"github.com/XingchengZhu/sms-webhook/sms"
)

func main() {
	cfg := config.LoadConfig()

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(cfg.LogLevel)

	// 老的单通道兜底 sender（如果后面没配 SMS_PROVIDERS_JSON，就用这个）
	fallback := buildFallbackSender(cfg)

	// 多通道 manager：默认模式是 pick，只发一条
	mgr := sms.NewManager(
		cfg.SMSProvidersJSON,
		fallback,
		cfg.SMSTarget,
		cfg.SMSSendMode,
	)

	http.HandleFunc("/webhook", handlers.WebhookHandler(cfg, mgr))

	logrus.WithFields(logrus.Fields{
		"port":     cfg.Port,
		"mode":     cfg.SMSSendMode,
		"sms_api":  cfg.SMSAPIURL,
	}).Info("Starting webhook server")

	logrus.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}

func buildFallbackSender(cfg config.Config) sms.Sender {
	switch cfg.SMSProvider {
	case "form":
		return sms.NewFormSender(
			"default",
			cfg.SMSAPIURL,
			"code",
			"target",
			"content",
			cfg.SMSCode,
		)
	case "header-json":
		return sms.NewHeaderJSONSender(
			"default",
			cfg.SMSAPIURL,
			cfg.SMSCode,
			cfg.SMSAPIKey,
			cfg.SMSHeaderKey,
		)
	default: // json
		return sms.NewJSONSender(
			"default",
			cfg.SMSAPIURL,
			cfg.SMSCode,
		)
	}
}
