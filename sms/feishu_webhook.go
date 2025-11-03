// sms/feishu_webhook.go
package sms

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type FeishuWebhookSender struct {
	WebhookURL string
	Secret     string // 可空
	Client     *http.Client
}

func (f *FeishuWebhookSender) Send(_ string, content string) error {
	if f.Client == nil {
		f.Client = &http.Client{Timeout: 10 * time.Second}
	}

	payload := map[string]any{
		"msg_type": "text",
		"content":  map[string]string{"text": content},
	}

	// 开启签名校验时，飞书要求把 timestamp 与 sign 放到 body 顶层字段
	if f.Secret != "" {
		ts := fmt.Sprintf("%d", time.Now().Unix())
		sign := genWebhookSign(ts, f.Secret)
		payload["timestamp"] = ts
		payload["sign"] = sign
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, f.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := f.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 2xx 视为成功；飞书会返回 {"StatusCode":0,...}，这里不强依赖解析
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("feishu webhook status=%d", resp.StatusCode)
	}
	logrus.WithField("feishu_webhook", f.WebhookURL).Debug("sent to feishu webhook")
	return nil
}

// sign = Base64(HMAC-SHA256(timestamp + "\n" + secret)) 参考官方“自定义机器人”文档
func genWebhookSign(ts, secret string) string {
	stringToSign := ts + "\n" + secret
	mac := hmac.New(sha256.New, []byte(stringToSign))
	// 注意算法要求对空串做 HMAC
	_, _ = mac.Write(nil)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
