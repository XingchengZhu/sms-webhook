package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type HeaderJSONSender struct {
	name      string
	URL       string
	Code      string
	APIKey    string
	HeaderKey string
	Client    *http.Client
}

type headerJSONPayload struct {
	Code    string `json:"code"`
	Mobile  string `json:"mobile"`
	Content string `json:"content"`
}

func NewHeaderJSONSender(name, urlStr, code, apiKey, headerKey string) *HeaderJSONSender {
	if name == "" {
		name = "header"
	}
	if headerKey == "" {
		headerKey = "X-API-KEY"
	}
	return &HeaderJSONSender{
		name:      name,
		URL:       urlStr,
		Code:      code,
		APIKey:    apiKey,
		HeaderKey: headerKey,
	}
}

func (s *HeaderJSONSender) Name() string {
	return s.name
}

func (s *HeaderJSONSender) Send(target, content string) error {
	payload := headerJSONPayload{
		Code:    s.Code,
		Mobile:  target,
		Content: content,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.URL, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.APIKey != "" && s.HeaderKey != "" {
		req.Header.Set(s.HeaderKey, s.APIKey)
	}

	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	logrus.WithFields(logrus.Fields{
		"sender": "header",
		"status": resp.StatusCode,
		"resp":   string(respBody),
	}).Info("header json sms response")

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("header json sms: unexpected status %d", resp.StatusCode)
	}

	return nil
}
