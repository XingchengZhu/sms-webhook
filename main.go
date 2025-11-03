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

    // 兼容：如果只给了一条 SMS_API_URL，就做成一个默认 sender
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

    logrus.WithFields(logrus.Fields{
        "port": cfg.Port,
    }).Info("server starting")

    logrus.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
