package config

import (
    "os"

    "github.com/sirupsen/logrus"
)

type Config struct {
    Port         string
    LogLevel     logrus.Level

    // 兼容单通道的老配置
    SMSAPIURL    string
    SMSCode      string
    SMSTarget    string

    // 多通道的 JSON 配置
    SMSProvidersJSON string // 环境变量：SMS_PROVIDERS_JSON

    // 发法：pick（默认）/ broadcast
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
        // 这里把默认从 broadcast 改成 pick
        SMSSendMode:      getEnv("SMS_SEND_MODE", "pick"),
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
