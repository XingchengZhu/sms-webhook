package config

import (
    "os"

    "github.com/sirupsen/logrus"
)

type Config struct {
    SMSAPIURL    string
    SMSCode      string
    SMSTarget    string
    Port         string
    LogLevel     logrus.Level
    SMSProvider  string // json | form | header-json
    SMSAPIKey    string // for header-json
    SMSHeaderKey string // for header-json
}

func LoadConfig() Config {
    return Config{
        SMSAPIURL:    getEnv("SMS_API_URL", "http://fake-sms.sms-webhook.svc.cluster.local:9999/sms"),
        SMSCode:      getEnv("SMS_CODE", "ALERT_CODE"),
        SMSTarget:    getEnv("SMS_TARGET", "15222222222"),
        Port:         getEnv("PORT", "8080"),
        LogLevel:     getLogLevel(getEnv("LOG_LEVEL", "info")),
        SMSProvider:  getEnv("SMS_PROVIDER", "json"),
        SMSAPIKey:    getEnv("SMS_API_KEY", ""),
        SMSHeaderKey: getEnv("SMS_HEADER_KEY", "X-API-KEY"),
    }
}

func getEnv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func getLogLevel(level string) logrus.Level {
    lvl, err := logrus.ParseLevel(level)
    if err != nil {
        return logrus.InfoLevel
    }
    return lvl
}
