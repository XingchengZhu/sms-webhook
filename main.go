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

    // fallback：为了兼容你现在“只有一条短信配置”的情况
    var fallback sms.Sender
    if cfg.SMSAPIURL != "" {
        fallback = sms.NewJSONSender("default", cfg.SMSAPIURL, cfg.SMSCode)
    }

    manager := sms.NewManager(
        cfg.SMSProvidersJSON, // 多通道 JSON
        fallback,             // 没有多通道时就用单通道
        cfg.SMSTarget,
        cfg.SMSSendMode,
    )

    http.HandleFunc("/webhook", handlers.WebhookHandler(cfg, manager))

    logrus.WithFields(logrus.Fields{
        "port": cfg.Port,
    }).Info("server starting")

    logrus.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
