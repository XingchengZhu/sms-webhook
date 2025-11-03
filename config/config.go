package config

import (
    "os"

    "github.com/sirupsen/logrus"
)

type Config struct {
    Port         string
    LogLevel     logrus.Level

    // 老的单通道配置（保持兼容）
    SMSAPIURL    string
    SMSCode      string
    SMSTarget    string

    // 新的多通道配置（可选）
    SMSProvidersJSON string // JSON 字符串，数组
    // 发送模式：broadcast | pick | both
    SMSSendMode      string
}

func LoadConfig() Config {
    return Config{
        Port:             getEnv("PORT", "8080"),
        LogLevel:         getLogLevel(getEnv("LOG_LEVEL", "info")),
        SMSAPIURL:        getEnv("SMS_API_URL", ""),
        SMSCode:          getEnv("SMS_CODE", ""),
        SMSTarget:        getEnv("SMS_TARGET", ""),
        SMSProvidersJSON: getEnv("SMS_PROVIDERS_JSON", ""),   // 新的
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
