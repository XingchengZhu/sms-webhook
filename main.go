package main

import (
    "net/http"

    "github.com/sirupsen/logrus"

    "github.com/XingchengZhu/sms-webhook/config"
    "github.com/XingchengZhu/sms-webhook/handlers"
    "github.com/XingchengZhu/sms-webhook/sms"
)

func main() {
    cfg := config.Load()

    log := logrus.New()
    level, err := logrus.ParseLevel(cfg.LogLevel)
    if err != nil {
        level = logrus.InfoLevel
    }
    log.SetLevel(level)

    sender := sms.NewSender(cfg, log)

    mux := http.NewServeMux()
    mux.Handle("/webhook", handlers.NewWebhookHandler(sender, cfg, log))

    log.WithFields(logrus.Fields{
        "port":     cfg.Port,
        "channels": len(cfg.Channels),
    }).Info("Starting webhook server")

    if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
        log.Fatal(err)
    }
}
