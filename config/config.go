package config

import (
    "os"

    "github.com/sirupsen/logrus"
)

type Config struct {
    Port         string
    LogLevel     logrus.Level

    // 老的单通道配置（兼容现在这版）
    SMSAPIURL    string
    SMSCode      string
    SMSTarget    string

    // 新的：一次性配多条通道，用 JSON
    SMSProvidersJSON string // 环境变量：SMS_PROVIDERS_JSON

    // 发法：broadcast（默认）/ pick
    SMSSendMode      string
}

func LoadConfig() Config {
    return Config{
        Port:             getEnv("PORT", "8080"),
        LogLevel:         getLogLevel(getEnv("LOG_LEVEL", "info")),
        SMSAPIURL:        getEnv("SMS_API_URL", ""),
        SMSCode:          getEnv("SMS_CODE", ""),
        SMSTarget:        getEnv("SMS_TARGET", ""),
        SMSProvidersJSON: getEnv("SMS_PROVIDERS_JSON", ""),
        SMSSendMode:      getEnv("SMS_SEND_MODE", "broadcast"),
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
