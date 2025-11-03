// sms/feishu_api.go
package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type FeishuAPISender struct {
	AppID     string
	AppSecret string

	ReceiveIDType string // chat_id | open_id | user_id | email
	ReceiveID     string // oc_xxx / ou_xxx / 邮箱

	Client *http.Client

	mu          sync.Mutex
	token       string
	tokenExpire time.Time
	ttl         time.Duration // 默认 2h
}

func (s *FeishuAPISender) Send(target, content string) error {
	if s.Client == nil {
		s.Client = &http.Client{Timeout: 10 * time.Second}
	}
	if s.ttl == 0 {
		s.ttl = 2 * time.Hour
	}

	// 允许用调用时 target 覆盖配置
	receiveID := s.ReceiveID
	if target != "" {
		receiveID = target
	}
	if receiveID == "" || s.ReceiveIDType == "" {
		return fmt.Errorf("feishu api receive_id/receive_id_type required")
	}

	token, err := s.ensureToken()
	if err != nil {
		return err
	}

	endpoint := "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=" + s.ReceiveIDType
	body := map[string]any{
		"receive_id": receiveID,
		"msg_type":   "text",
		"content":    mustJSON(map[string]string{"text": content}), // 注意 content 需要是字符串化的 JSON
	}
	bs, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(bs))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("feishu api status=%d", resp.StatusCode)
	}
	return nil
}

func (s *FeishuAPISender) ensureToken() (string, error) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token != "" && now.Before(s.tokenExpire.Add(-1*time.Minute)) {
		return s.token, nil
	}

	type tokenResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Tok  string `json:"tenant_access_token"`
		Exp  int    `json:"expire"` // 秒
	}
	payload := map[string]string{
		"app_id":     s.AppID,
		"app_secret": s.AppSecret,
	}
	bs, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal", bytes.NewReader(bs))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := s.http().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("get tenant_access_token status=%d", resp.StatusCode)
	}

	var tr tokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}
	if tr.Code != 0 || tr.Tok == "" {
		return "", fmt.Errorf("get tenant_access_token failed: code=%d msg=%s", tr.Code, tr.Msg)
	}

	s.token = tr.Tok
	// 官方过期时间通常是 7200 秒；这里以返回为准
	exp := time.Duration(tr.Exp) * time.Second
	if exp <= 0 {
		exp = s.ttl
	}
	s.tokenExpire = time.Now().Add(exp)
	return s.token, nil
}

func (s *FeishuAPISender) http() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	return &http.Client{Timeout: 10 * time.Second}
}

func mustJSON(v any) string {
	bs, _ := json.Marshal(v)
	return string(bs)
}
