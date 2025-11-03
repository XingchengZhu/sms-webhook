package config

import (
    "encoding/json"
    "os"
    "strings"
)

type SMSChannel struct {
    Name       string            `json:"name"`        // 通道名，比如 json、form1、aliyun
    Type       string            `json:"type"`        // json | form | header | text
    URL        string            `json:"url"`         // 要打的地址
    Method     string            `json:"method"`      // POST / GET, 默认 POST
    CodeKey    string            `json:"code_key"`    // form/json 时的字段名，默认 code
    TargetKey  string            `json:"target_key"`  // form/json 时的字段名，默认 target
    ContentKey string            `json:"content_key"` // form/json 时的字段名，默认 content
    Static     map[string]string `json:"static"`      // 额外要带的 kv
    Headers    map[string]string `json:"headers"`     // 额外 header
}

type Config struct {
    Port            string
    LogLevel        string
    Channels        map[string]SMSChannel
    DefaultChannels []string
    Broadcast       bool
    DefaultCode     string
    DefaultTarget   string
}

func getEnv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func Load() Config {
    port := getEnv("PORT", "8080")
    logLevel := getEnv("LOG_LEVEL", "info")
    broadcast := getEnv("SMS_BROADCAST", "true") == "true"

    cfg := Config{
        Port:          port,
        LogLevel:      logLevel,
        Channels:      make(map[string]SMSChannel),
        Broadcast:     broadcast,
        DefaultCode:   getEnv("SMS_CODE", "ALERT_CODE"),
        DefaultTarget: getEnv("SMS_TARGET", "15222222222"),
    }

    // 1) 新的多通道写法：SMS_CHANNELS 是一段 JSON
    if raw := os.Getenv("SMS_CHANNELS"); raw != "" {
        var list []SMSChannel
        if err := json.Unmarshal([]byte(raw), &list); err == nil {
            for _, ch := range list {
                if ch.Method == "" {
                    ch.Method = "POST"
                }
                if ch.Type == "" {
                    ch.Type = "json"
                }
                if ch.CodeKey == "" {
                    ch.CodeKey = "code"
                }
                if ch.TargetKey == "" {
                    ch.TargetKey = "target"
                }
                if ch.ContentKey == "" {
                    ch.ContentKey = "content"
                }
                if ch.Static == nil {
                    ch.Static = map[string]string{}
                }
                if ch.Headers == nil {
                    ch.Headers = map[string]string{}
                }
                cfg.Channels[ch.Name] = ch
            }
        }
    } else {
        // 2) 兼容你原来那种“只有一个短信接口”的写法
        single := SMSChannel{
            Name:   "default",
            Type:   getEnv("SMS_PROVIDER", "json"),
            URL:    getEnv("SMS_API_URL", "http://127.0.0.1:9999/sms"),
            Method: "POST",
        }
        cfg.Channels["default"] = single
    }

    // 默认通道列表
    defCh := getEnv("DEFAULT_CHANNELS", "")
    if defCh != "" {
        parts := strings.Split(defCh, ",")
        for _, p := range parts {
            p = strings.TrimSpace(p)
            if p != "" {
                cfg.DefaultChannels = append(cfg.DefaultChannels, p)
            }
        }
    }

    return cfg
}
