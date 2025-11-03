package sms

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// FeishuSender 通过飞书机器人 Webhook 发送文本消息
// 说明：开启“签名校验”后，需要在请求体携带 timestamp 与 sign
// sign = Base64( HMAC-SHA256( secret, timestamp + "\n" + secret ) )
type FeishuSender struct {
	name      string
	webhook   string
	secret    string
	httpc     *http.Client
	userAgent string
}

func NewFeishuSender(name, webhook, secret string) *FeishuSender {
	if name == "" {
		name = "feishu"
	}
	return &FeishuSender{
		name:      name,
		webhook:   webhook,
		secret:    secret,
		httpc:     &http.Client{Timeout: 8 * time.Second},
		userAgent: "sms-webhook/feishu-sender",
	}
}

func (s *FeishuSender) Name() string { return s.name }

// Send 忽略 target，只使用 content
func (s *FeishuSender) Send(_ string, content string) error {
	if s.webhook == "" {
		return fmt.Errorf("feishu webhook is empty")
	}

	ts := fmt.Sprintf("%d", time.Now().Unix())
	sign := ""
	if s.secret != "" {
		sign = signFeishu(ts, s.secret)
	}

	// 飞书 text 消息体
	body := map[string]any{
		"msg_type": "text",
		"content": map[string]string{
			"text": content,
		},
	}
	// 若开启签名校验，需要同时传 timestamp 与 sign
	if s.secret != "" {
		body["timestamp"] = ts
		body["sign"] = sign
	}

	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, s.webhook, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", s.userAgent)

	resp, err := s.httpc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rb, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("feishu http %d: %s", resp.StatusCode, string(rb))
	}

	// 飞书返回通常包含 code/statusCode == 0 表示成功
	var r struct {
		Code        *int   `json:"code"`
		Msg         string `json:"msg"`
		StatusCode  *int   `json:"StatusCode"`
		StatusMsg   string `json:"StatusMessage"`
		Extra       any    `json:"Extra"`
	}
	_ = json.Unmarshal(rb, &r)
	ok := false
	if r.Code != nil && *r.Code == 0 {
		ok = true
	}
	if r.StatusCode != nil && *r.StatusCode == 0 {
		ok = true
	}
	logrus.WithFields(logrus.Fields{
		"sender": s.name,
		"resp":   string(rb),
		"url":    s.webhook,
	}).Info("feishu response")

	if !ok {
		return fmt.Errorf("feishu response not ok: %s", string(rb))
	}
	return nil
}

func signFeishu(ts, secret string) string {
	// sign = Base64(HMAC-SHA256(secret, ts+"\n"+secret))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "\n" + secret))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
