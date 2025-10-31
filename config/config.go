package config

import (
	"os"
	"github.com/sirupsen/logrus"
)

type Config struct {
	SMSAPIURL string
	SMSCode   string
	SMSTarget string
	Port      string
	LogLevel  logrus.Level
}

func LoadConfig() Config {
	return Config{
		SMSAPIURL: getEnv("SMS_API_URL", "https://default-sms-api-url.com/send"),
		SMSCode:   getEnv("SMS_CODE", "ALERT_CODE"),
		SMSTarget: getEnv("SMS_TARGET", "15222222222"),
		Port:      getEnv("PORT", "8080"),
		LogLevel:  getLogLevel(getEnv("LOG_LEVEL", "info")),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getLogLevel(level string) logrus.Level {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return logrus.InfoLevel
	}
	return lvl
}
