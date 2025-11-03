package config

import (
	"os"

	"github.com/sirupsen/logrus"
)

type Config struct {
	// http
	Port     string
	LogLevel logrus.Level

	// 老的单通道写法
	SMSAPIURL    string
	SMSCode      string
	SMSTarget    string
	SMSProvider  string // json | form | header-json
	SMSAPIKey    string
	SMSHeaderKey string

	// 新的多通道写法
	SMSSendMode      string // pick(默认) | broadcast
	SMSProvidersJSON string // JSON 数组
}

func LoadConfig() Config {
	return Config{
		Port:     getEnv("PORT", "8080"),
		LogLevel: getLogLevel(getEnv("LOG_LEVEL", "info")),

		// 老的单通道
		SMSAPIURL:    getEnv("SMS_API_URL", "http://fake-sms.sms-webhook.svc.cluster.local:9999/json"),
		SMSCode:      getEnv("SMS_CODE", "ALERT_CODE"),
		SMSTarget:    getEnv("SMS_TARGET", "15222222222"),
		SMSProvider:  getEnv("SMS_PROVIDER", "json"),
		SMSAPIKey:    getEnv("SMS_API_KEY", ""),
		SMSHeaderKey: getEnv("SMS_HEADER_KEY", "X-API-KEY"),

		// 新的多通道，默认就是只发一条
		SMSSendMode:      getEnv("SMS_SEND_MODE", "pick"),
		SMSProvidersJSON: getEnv("SMS_PROVIDERS_JSON", ""),
	}
}

func getEnv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func getLogLevel(level string) logrus.Level {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		return logrus.InfoLevel
	}
	return l
}
