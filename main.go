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

    // 单通道 fallback（兼容你现在这版）
    var fallback sms.Sender
    if cfg.SMSAPIURL != "" {
        fallback = sms.NewJSONSender("default", cfg.SMSAPIURL, cfg.SMSCode)
    }

    manager := sms.NewManager(
        cfg.SMSProvidersJSON,
        fallback,
        cfg.SMSTarget,
        cfg.SMSSendMode,
    )

    http.HandleFunc("/webhook", handlers.WebhookHandler(cfg, manager))

    logrus.WithField("port", cfg.Port).Info("server starting")
    logrus.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
