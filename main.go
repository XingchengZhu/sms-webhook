// main.go
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

    // 根据配置创建一个 sender
    sender := buildSender(cfg)

    http.HandleFunc("/webhook", handlers.WebhookHandler(cfg, sender))

    logrus.WithFields(logrus.Fields{
        "port":      cfg.Port,
        "provider":  cfg.SMSProvider,
        "sms_api":   cfg.SMSAPIURL,
    }).Info("Starting webhook server")

    logrus.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}

func buildSender(cfg config.Config) sms.Sender {
    switch cfg.SMSProvider {
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
    default: // "json"
        return &sms.JSONSender{
            URL:  cfg.SMSAPIURL,
            Code: cfg.SMSCode,
        }
    }
}
