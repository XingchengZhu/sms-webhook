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

// FeishuSender implements Sender to push messages via Feishu (Lark) custom bot webhook.
type FeishuSender struct {
	name   string
	URL    string
	Secret string // optional: if set, will add timestamp & sign fields
	// Message type: "text" (default). Could be extended to "post" / "interactive".
	MsgType string
}

func NewFeishuSender(name, url, secret string) *FeishuSender {
	return &FeishuSender{
		name:    name,
		URL:     url,
		Secret:  secret,
		MsgType: "text",
	}
}

func (s *FeishuSender) Name() string { return s.name }

type feishuTextPayload struct {
	MsgType   string            `json:"msg_type"`
	Content   map[string]string `json:"content"`
	Timestamp string            `json:"timestamp,omitempty"`
	Sign      string            `json:"sign,omitempty"`
}

func (s *FeishuSender) Send(_target, content string) error {
	// Build payload
	payload := feishuTextPayload{
		MsgType: s.MsgType,
		Content: map[string]string{
			"text": content,
		},
	}

	// 如果开启“签名校验”，需要在 body 里带 timestamp & sign
	// 签名算法（官方文档）：sign = Base64( HMAC-SHA256( key = timestamp + "\n" + secret, msg = "" ) )
	// 并且 timestamp 单位为秒，误差不超过 1 小时。
	// 参考：飞书自定义机器人指南（签名说明与 body 字段）:contentReference[oaicite:0]{index=0}
	if s.Secret != "" {
		ts := time.Now().Unix()
		sign, err := feishuSign(s.Secret, ts)
		if err != nil {
			return fmt.Errorf("feishu sign failed: %w", err)
		}
		payload.Timestamp = fmt.Sprintf("%d", ts)
		payload.Sign = sign
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", s.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	logrus.WithFields(logrus.Fields{
		"sender": "feishu-webhook",
		"status": resp.StatusCode,
		"resp":   string(respBody),
	}).Info("feishu webhook response")

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("feishu webhook: unexpected status %d", resp.StatusCode)
	}

	// 官方返回通常是 200 且 JSON: {"StatusCode":0,"StatusMessage":"success",...}
	// 这里做个兜底校验（不强依赖）
	type feishuResp struct {
		StatusCode    int    `json:"StatusCode"`
		StatusMessage string `json:"StatusMessage"`
	}
	var fr feishuResp
	if err := json.Unmarshal(respBody, &fr); err == nil && fr.StatusCode != 0 {
		return fmt.Errorf("feishu webhook: code=%d message=%s", fr.StatusCode, fr.StatusMessage)
	}

	return nil
}

func feishuSign(secret string, ts int64) (string, error) {
	key := []byte(fmt.Sprintf("%d\n%s", ts, secret))
	m := hmac.New(sha256.New, key)
	// message is empty string
	if _, err := m.Write([]byte("")); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(m.Sum(nil)), nil
}
