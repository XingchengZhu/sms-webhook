package main

import (
	"net/http"
	"sms-webhook/config"
	"sms-webhook/handlers"

	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.LoadConfig()

	// 设置日志等级
	logrus.SetLevel(cfg.LogLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	logrus.WithFields(logrus.Fields{
		"sms_api_url": cfg.SMSAPIURL,
		"port":        cfg.Port,
	}).Info("Starting webhook server")

	http.HandleFunc("/webhook", handlers.WebhookHandler(cfg))
	logrus.Infof("Server is running on port %s", cfg.Port)
	logrus.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
