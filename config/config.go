// config/config.go
package config

import (
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Config struct {
	// 原有
	SMSAPIURL    string
	SMSCode      string
	SMSTarget    string
	Port         string
	LogLevel     logrus.Level
	SMSProvider  string // 单路：json | form | header-json | feishu-webhook | feishu-api | pick | broadcast
	SMSAPIKey    string // for header-json
	SMSHeaderKey string // for header-json

	// 多路组合
	SMSProviders []string // 逗号分隔: e.g. "feishu-webhook,json"

	// 飞书 Webhook
	FeishuWebhook string // e.g. https://open.feishu.cn/open-apis/bot/v2/hook/xxx
	FeishuSecret  string // 开启签名校验时配置；可空

	// 飞书 API（应用）
	FeishuAppID          string
	FeishuAppSecret      string
	FeishuReceiveIDType  string // chat_id | open_id | user_id | email
	FeishuReceiveID      string // oc_xxx / ou_xxx / 用户邮箱等
	FeishuTokenCacheSecs int    // 可选，默认7200（2h）
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

		SMSProviders: splitAndTrim(getEnv("SMS_PROVIDERS", "")),

		FeishuWebhook:        getEnv("FEISHU_WEBHOOK", ""),
		FeishuSecret:         getEnv("FEISHU_SECRET", ""),
		FeishuAppID:          getEnv("FEISHU_APP_ID", ""),
		FeishuAppSecret:      getEnv("FEISHU_APP_SECRET", ""),
		FeishuReceiveIDType:  getEnv("FEISHU_RECEIVE_ID_TYPE", "chat_id"),
		FeishuReceiveID:      getEnv("FEISHU_RECEIVE_ID", ""),
		FeishuTokenCacheSecs: getEnvInt("FEISHU_TOKEN_CACHE_SECS", 7200),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var x int
		_, _ = fmt.Sscanf(v, "%d", &x)
		if x > 0 {
			return x
		}
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

func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// 帮助：为飞书 token 过期计算提供统一 now()
func Now() time.Time { return time.Now() }
