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

// FeishuSender 发送飞书自定义机器人消息（v2 hook + 签名）
type FeishuSender struct {
	name       string
	webhookURL string
	secret     string
	client     *http.Client
}

func NewFeishuSender(name, webhookURL, secret string) Sender {
	if webhookURL == "" || secret == "" {
		logrus.WithFields(logrus.Fields{
			"name":   name,
			"hasURL": webhookURL != "",
			"hasSec": secret != "",
		}).Warn("feishu sender init skipped: missing url or secret")
		return nil
	}
	if name == "" {
		name = "feishu"
	}
	return &FeishuSender{
		name:       name,
		webhookURL: webhookURL,
		secret:     secret,
		client:     &http.Client{Timeout: 8 * time.Second},
	}
}

func (s *FeishuSender) Name() string { return s.name }

func (s *FeishuSender) Send(target, content string) error {
	ts := fmt.Sprintf("%d", time.Now().Unix())
	sign := s.sign(ts)

	// 把 target 也拼到文本里，便于审计（飞书机器人本身不需要手机号）
	text := content
	if target != "" {
		text = fmt.Sprintf("%s\n(target: %s)", content, target)
	}

	payload := map[string]any{
		"timestamp": ts,
		"sign":      sign,
		"msg_type":  "text",
		"content": map[string]string{
			"text": text,
		},
	}
	bs, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", s.webhookURL, bytes.NewReader(bs))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("feishu http %d: %s", resp.StatusCode, string(body))
	}

	// 飞书机器人成功通常 200 + {"StatusCode":0,"StatusMessage":"success"} 或 {"code":0,"msg":"ok"}
	var r map[string]any
	if err := json.Unmarshal(body, &r); err == nil {
		if code, ok := r["StatusCode"]; ok {
			if iv, isNum := code.(float64); isNum && iv != 0 {
				return fmt.Errorf("feishu response not ok: %v", r)
			}
		}
		if code, ok := r["code"]; ok {
			if iv, isNum := code.(float64); isNum && iv != 0 {
				return fmt.Errorf("feishu response not ok: %v", r)
			}
		}
	}

	logrus.WithField("sender", s.name).Info("feishu message sent")
	return nil
}

// 飞书签名：Base64( HMAC-SHA256( timestamp + "\n" + secret , key=secret ) )
func (s *FeishuSender) sign(ts string) string {
	data := ts + "\n" + s.secret
	mac := hmac.New(sha256.New, []byte(s.secret))
	_, _ = mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
